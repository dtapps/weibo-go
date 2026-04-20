package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/types"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.SetLevel(logger.LevelError)
}

// mockCallback 模拟回调
type mockCallback struct {
	mu            sync.Mutex
	onReadyCalled bool
	stateHistory  []string
}

func (m *mockCallback) OnReady(data *types.OnReadyData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onReadyCalled = true
}

func (m *mockCallback) OnStateChange(state string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stateHistory = append(m.stateHistory, state)
}

func (m *mockCallback) OnDispatch(msg *types.WsMessageMsg)               {}
func (m *mockCallback) OnError(err error)                                {}
func (m *mockCallback) OnClose(code int, reason string)                  {}
func (m *mockCallback) OnKickout(code int, reason string)                {}
func (m *mockCallback) OnAuthFailed(code int) (*types.WsAuthData, error) { return nil, nil }

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// TestWsClient_ConcurrentStateAccess 专门压力测试 RWMutex
// 验证在频繁获取状态和断开连接时，是否会发生竞态或死锁
func TestWsClient_ConcurrentStateAccess(t *testing.T) {
	cb := &mockCallback{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		conn, _ := upgrader.Upgrade(w, nil, nil)
		_ = conn.Close()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewWsClient(wsURL, "race_test", cb)

	// 启动连接
	go func() {
		_ = client.Connect()
	}()

	// 并发读取状态和发送消息
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			client.GetState()
		}()
		go func() {
			defer wg.Done()
			_ = client.SendMessage("123", "test", "mid", 0, true)
		}()
	}

	time.Sleep(100 * time.Millisecond)
	_ = client.Disconnect()
	wg.Wait()
}

// TestWsClient_ReconnectLogic 验证重连定时器与锁的配合
func TestWsClient_ReconnectLogic(t *testing.T) {
	cb := &mockCallback{}
	var connCount int
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		connCount++
		mu.Unlock()

		conn, _ := upgrader.Upgrade(w, nil, nil)
		// 模拟服务器立即断开
		_ = conn.Close()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewWsClient(wsURL, "reconnect_test", cb)

	// 设置极短重连：10ms, 20ms
	client.SetReconnectConfig(2, "10ms,20ms")

	_ = client.Connect()

	// 给够 3 次连接的时间（初始1次 + 重连2次）
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	finalCount := connCount
	mu.Unlock()

	// 验证重连次数是否符合预期（不应超过 maxReconnectAttempts）
	assert.LessOrEqual(t, finalCount, 3)
}

// TestWsClient_HeartbeatSafety 测试心跳在连接断开时的安全性
func TestWsClient_HeartbeatSafety(t *testing.T) {
	cb := &mockCallback{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		conn, _ := upgrader.Upgrade(w, nil, nil)
		time.Sleep(50 * time.Millisecond)
		_ = conn.Close()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewWsClient(wsURL, "hb_test", cb)
	client.mu.Lock()
	client.heartbeatInterval = 20 * time.Millisecond // 快速心跳
	client.mu.Unlock()

	_ = client.Connect()

	// 在心跳进行中突然断开
	time.Sleep(50 * time.Millisecond)
	err := client.Disconnect()
	assert.NoError(t, err)

	// 确保断开后心跳定时器不再运行，不会触发 nil 指针或已关闭连接的写入
	time.Sleep(100 * time.Millisecond)
}
