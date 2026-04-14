package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/types"
	"github.com/gorilla/websocket"
)

const (
	// 默认心跳间隔(秒)
	DefaultHeartbeatInterval = 30
	// 默认心跳超时次数阈值
	HeartbeatTimeoutThreshold = 2
	// 默认发送超时(毫秒)
	DefaultSendTimeoutMs = 30000
	// 默认最大重连次数
	DefaultMaxReconnectAttempts = 100
	// 默认重连延迟
	DefaultReconnectDelays = "1s,2s,5s,10s,30s,60s"
)

// WsClientCallback WebSocket客户端回调接口
type WsClientCallback interface {
	OnReady(data *types.OnReadyData)
	OnDispatch(pushEvent *types.WsMessageMsg)
	OnStateChange(state string)
	OnError(err error)
	OnClose(code int, reason string)
	OnKickout(code int, reason string)
	OnAuthFailed(code int) (*types.WsAuth, error)
}

type WsClient struct {
	mu        sync.RWMutex
	conn      *websocket.Conn
	url       string
	state     string
	auth      *types.WsAuth
	accountId string
	botId     string

	// 心跳相关
	heartbeatInterval     int         // 心跳间隔(秒)
	heartbeatTimer        *time.Timer // 心跳定时器
	heartbeatAckReceived  bool        // 是否收到心跳确认
	lastHeartbeatAt       int64       // 上次心跳时间
	heartbeatTimeoutCount int         // 心跳超时次数
	heartbeatCount        int         // 心跳次数

	// 重连相关
	reconnectAttempts    int
	maxReconnectAttempts int
	reconnectDelays      []time.Duration
	reconnectTimer       *time.Timer

	// 待响应的请求
	pendingRequests map[string]*PendingRequest

	// 回调
	callback WsClientCallback

	// 日志
	log *logger.Logger

	// 序列号
	seqNo uint32

	// 上下文
	ctx    context.Context
	cancel context.CancelFunc

	// 连接ID
	connectId string
}

// PendingRequest 待响应的请求
type PendingRequest struct {
	resolveCh chan any
	timeout   time.Duration
	decoder   func(data []byte, msgId string) any
	Resolve   func(any)
	Reject    func(error)
	Timer     *time.Timer
	Decoder   func([]byte, string) any
}

// NewWsClient 创建WebSocket客户端
func NewWsClient(url string, accountId string, botId string, callback WsClientCallback) *WsClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &WsClient{
		url:                  url,
		accountId:            accountId,
		botId:                botId,
		state:                "disconnected",
		heartbeatInterval:    DefaultHeartbeatInterval,
		maxReconnectAttempts: DefaultMaxReconnectAttempts,
		reconnectDelays:      parseDelays(DefaultReconnectDelays),
		pendingRequests:      make(map[string]*PendingRequest),
		callback:             callback,
		log:                  logger.New("ws"),
		ctx:                  ctx,
		cancel:               cancel,
	}
}

// SetAuth 设置认证信息
func (c *WsClient) SetAuth(auth *types.WsAuth) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.auth = auth
}

// SetReconnectConfig 设置重连配置
func (c *WsClient) SetReconnectConfig(maxAttempts int, delays string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxReconnectAttempts = maxAttempts
	c.reconnectDelays = parseDelays(delays)
}

// Connect 连接
func (c *WsClient) Connect() error {
	c.mu.Lock()
	if c.state == "disconnected" {
		c.state = "connecting"
	}
	c.mu.Unlock()

	return c.doConnect()
}

// doConnect 执行连接
func (c *WsClient) doConnect() error {

	c.mu.RLock()
	auth := c.auth
	c.mu.RUnlock()

	url := c.url
	if auth != nil && auth.Token != "" {
		url = fmt.Sprintf("%s?app_id=%s&token=%s&version=%s",
			c.url, auth.UID, auth.Token, "1.0.0")
	}

	c.log.Info("正在连接WebSocket", logger.F("url", url))

	// 创建WebSocket连接
	header := http.Header{}
	header.Set("Origin", c.url)

	conn, resp, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		c.log.Error("WebSocket连接失败", logger.F("error", err.Error()))
		c.scheduleReconnect()
		return err
	}

	if resp != nil {
		resp.Body.Close()
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	// 设置处理器
	conn.SetCloseHandler(func(code int, text string) error {
		c.handleClose(code, text)
		return nil
	})

	// 启动心跳定时器
	c.startHeartbeat()

	// 启动读取协程
	go c.readLoop()

	return nil
}

// readLoop 读取循环
func (c *WsClient) readLoop() {
	defer func() {
		// 连接断开时触发重连
		c.handleClose(1006, "read loop exit")
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:

			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				return
			}

			_, data, err := conn.ReadMessage()
			if err != nil {
				// 忽略 "use of closed" 错误，这是正常关闭
				if !strings.Contains(err.Error(), "use of closed") {
					c.log.Error("读取消息失败", logger.F("error", err.Error()))
				}
				return
			}

			c.log.Debug("收到原始消息",
				logger.F("length", len(data)),
				logger.F("data", string(data)),
			)

			c.handleMessage(data)
		}
	}
}

// handleMessage 处理消息
func (c *WsClient) handleMessage(data []byte) {
	c.mu.Lock()
	c.heartbeatAckReceived = true
	c.heartbeatTimeoutCount = 0
	c.mu.Unlock()

	text := string(data)

	// 提前 处理 pong 响应
	if text == "pong" || text == `{"type":"pong"}` {
		c.lastHeartbeatAt = time.Now().UnixMilli()
		return
	}

	var msg types.WsMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		c.log.Error("解析消息失败", logger.F("error", err.Error()))
		return
	}

	msgType := msg.Type
	c.log.Debug("收到消息", logger.F("msgType", msgType))

	switch msgType {
	case "connected":
		// 连接成功
		c.handlePushConnected(data)
	case "message":
		// 消息
		c.handlePushMessage(data)
	default:
		c.log.Debug("未处理的msgType", logger.F("msgType", msgType))
	}
}

// handleConnected 处理推送连接成功
func (c *WsClient) handlePushConnected(data []byte) {

	var msg types.WsConnectedMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		c.log.Error("解析消息失败", logger.F("error", err.Error()))
		return
	}

	c.mu.Lock()
	c.connectId = msg.ConnectionID
	c.state = "connected"
	c.mu.Unlock()

	// 回调
	if c.callback != nil {
		result := &types.OnReadyData{
			ConnectID: c.connectId,
			Timestamp: time.Now().Unix(),
		}
		c.callback.OnReady(result)
		c.callback.OnStateChange("connected")
	}
}

// handlePushMessage 处理推送消息
func (c *WsClient) handlePushMessage(data []byte) {
	c.log.Debug("收到推送消息", logger.F("data", string(data)))

	var msg types.WsMessageMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		c.log.Error("解析消息失败", logger.F("error", err.Error()))
		return
	}

	if c.callback != nil {
		c.callback.OnDispatch(&msg)
	}
}

// startHeartbeat 启动心跳
func (c *WsClient) startHeartbeat() {
	c.stopHeartbeat()

	c.mu.Lock()
	interval := time.Duration(c.heartbeatInterval) * time.Second
	c.mu.Unlock()

	c.heartbeatTimer = time.AfterFunc(interval, func() {
		c.sendHeartbeat()
	})
}

// scheduleReconnect 安排重连
func (c *WsClient) scheduleReconnect() {
	c.mu.Lock()
	if c.state == "reconnecting" {
		c.mu.Unlock()
		return
	}
	c.state = "reconnecting"
	c.mu.Unlock()

	c.stopHeartbeat()

	c.mu.Lock()
	c.reconnectAttempts++
	if c.maxReconnectAttempts > 0 && c.reconnectAttempts > c.maxReconnectAttempts {
		c.log.Error("达到最大重连次数", logger.F("attempts", c.reconnectAttempts))
		c.state = "disconnected"
		c.mu.Unlock()
		return
	}

	delay := c.getReconnectDelay()
	c.mu.Unlock()

	c.log.Info("计划重连", logger.F("delay", delay), logger.F("attempts", c.reconnectAttempts))

	c.reconnectTimer = time.AfterFunc(delay, func() {
		c.doConnect()
	})
}

// getReconnectDelay 获取重连延迟
func (c *WsClient) getReconnectDelay() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	idx := c.reconnectAttempts - 1
	if idx >= len(c.reconnectDelays) {
		idx = len(c.reconnectDelays) - 1
	}
	if idx < 0 {
		idx = 0
	}

	delay := c.reconnectDelays[idx]
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	return delay + jitter
}

// close 关闭连接
func (c *WsClient) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.stopHeartbeatLocked()

	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
		c.reconnectTimer = nil
	}
}

// handleClose 处理关闭
func (c *WsClient) handleClose(code int, reason string) {
	c.log.Info("WebSocket关闭", logger.F("code", code), logger.F("reason", reason))

	c.mu.Lock()
	wasConnected := c.state == "connected"
	c.mu.Unlock()

	c.stopHeartbeat()

	c.close()

	c.mu.Lock()
	c.state = "disconnected"
	c.mu.Unlock()

	if c.callback != nil {
		c.callback.OnClose(code, reason)
	}

	if wasConnected || code != 1000 {
		c.scheduleReconnect()
	}
}

// stopHeartbeatLocked 内部心跳停止逻辑 (加锁)
func (c *WsClient) stopHeartbeatLocked() {
	if c.heartbeatTimer != nil {
		c.heartbeatTimer.Stop()
		c.heartbeatTimer = nil
	}
}

// stopHeartbeat 对外暴露的心跳停止方法 (加锁)
func (c *WsClient) stopHeartbeat() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stopHeartbeatLocked()
}

// Disconnect 断开连接
func (c *WsClient) Disconnect() error {
	c.mu.Lock()
	c.state = "disconnected"
	c.mu.Unlock()

	c.cancel()
	c.close()

	return nil
}

// GetState 获取状态
func (c *WsClient) GetState() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// GetConnectId 获取连接ID
func (c *WsClient) GetConnectId() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connectId
}

func (c *WsClient) SendMessage(toUserId, text, messageId string, chunkId int, done bool) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("连接未建立")
	}

	msg := types.SendMessageRequest{
		Type: "send_message",
		Payload: types.SendMessagePayload{
			ToUserId:  toUserId,
			Text:      text,
			MessageId: messageId,
			ChunkId:   chunkId,
			Done:      done,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	return nil
}

// sendHeartbeat 发送心跳
func (c *WsClient) sendHeartbeat() {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return
	}

	c.log.Debug("准备发送心跳",
		logger.F("interval", c.heartbeatInterval),
		logger.F("lastAt", time.UnixMilli(c.lastHeartbeatAt).Format(time.DateTime)),
		logger.F("count", c.heartbeatTimeoutCount),
		logger.F("total", c.heartbeatCount),
	)

	now := time.Now().UnixMilli()
	c.mu.Lock()
	if now-c.lastHeartbeatAt > int64(c.heartbeatInterval*1000+c.heartbeatInterval*1000) {
		c.heartbeatTimeoutCount++
	}
	c.lastHeartbeatAt = now
	c.heartbeatCount++
	c.mu.Unlock()

	data, err := json.Marshal(types.HeartbeatMsgRequest{
		Type: "ping",
	})
	if err != nil {
		c.log.Error("序列化心跳失败", logger.F("error", err.Error()))
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		c.log.Error("发送心跳失败", logger.F("error", err.Error()))
	}

	c.log.Debug("发送心跳成功", logger.F("data", string(data)))

	c.mu.Lock()
	interval := time.Duration(c.heartbeatInterval) * time.Second
	c.mu.Unlock()

	c.heartbeatTimer = time.AfterFunc(interval, func() {
		c.sendHeartbeat()
	})
}

func parseDelays(delays string) []time.Duration {
	var result []time.Duration
	for _, d := range strings.Split(delays, ",") {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		var t time.Duration
		fmt.Sscanf(d, "%d", &t)
		result = append(result, t*time.Second)
	}
	if len(result) == 0 {
		result = []time.Duration{1 * time.Second}
	}
	return result
}
