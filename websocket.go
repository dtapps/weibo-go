package weibo

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pluginVersion = "2.0.0"
)

// WebSocketMessageHandler 消息处理器
type WebSocketMessageHandler func(data any)

// WebSocketErrorHandler 错误处理器
type WebSocketErrorHandler func(err error)

// WebSocketCloseHandler 关闭处理器
type WebSocketCloseHandler func(code int, reason string)

// WebSocketOpenHandler 打开处理器
type WebSocketOpenHandler func()

// WebSocketStatusHandler 状态处理器
type WebSocketStatusHandler func(patch *RuntimeStatusPatch)

// WebSocketClientOptions WebSocket 客户端选项
type WebSocketClientOptions struct {
	OnMessage            WebSocketMessageHandler
	OnError              WebSocketErrorHandler
	OnClose              WebSocketCloseHandler
	OnOpen               WebSocketOpenHandler
	OnStatus             WebSocketStatusHandler
	AutoReconnect        bool
	MaxReconnectAttempts int
}

// WebSocketClient WebSocket 客户端
type WebSocketClient struct {
	account      *ResolvedAccount
	options      *WebSocketClientOptions
	tokenManager *TokenManager

	ws *websocket.Conn
	mu sync.RWMutex

	// 心跳
	pingInterval *time.Timer
	lastPongTime int64
	pingTimeout  time.Duration

	// 重连
	reconnectAttempts int
	reconnectTimeout  *time.Timer
	isConnecting      bool
	shouldReconnect   bool
	closeChan         chan struct{}
}

// NewWebSocketClient 创建 WebSocket 客户端
func NewWebSocketClient(account *ResolvedAccount, options *WebSocketClientOptions, tokenManager *TokenManager) *WebSocketClient {
	if options == nil {
		options = &WebSocketClientOptions{
			AutoReconnect:        true,
			MaxReconnectAttempts: 0,
		}
	}

	return &WebSocketClient{
		account:         account,
		options:         options,
		tokenManager:    tokenManager,
		shouldReconnect: true,
		closeChan:       make(chan struct{}),
		pingTimeout:     PingTimeout,
	}
}

// Connect 连接
func (c *WebSocketClient) Connect() error {
	c.mu.Lock()
	if c.isConnecting {
		c.mu.Unlock()
		return nil
	}
	c.isConnecting = true
	c.shouldReconnect = true
	c.mu.Unlock()

	c.emitStatus(&RuntimeStatusPatch{
		Running:         boolPtr(true),
		Connected:       boolPtr(false),
		ConnectionState: ptrConnectionState(ConnectionStateConnecting),
		NextRetryAt:     nil,
		LastError:       nil,
	})

	if c.account.WSEndpoint == "" {
		c.isConnecting = false
		return fmt.Errorf("WebSocket endpoint not configured for account %s", c.account.AccountId)
	}

	if c.account.AppId == nil || *c.account.AppId == "" {
		c.isConnecting = false
		return fmt.Errorf("App ID not configured for account %s", c.account.AccountId)
	}

	appId := *c.account.AppId
	token, err := c.tokenManager.GetValidToken(
		*c.account.AppId,
		*c.account.AppSecret,
		c.account.TokenEndpoint,
	)
	if err != nil {
		c.isConnecting = false
		c.emitStatus(&RuntimeStatusPatch{
			Running:         boolPtr(true),
			Connected:       boolPtr(false),
			ConnectionState: ptrConnectionState(ConnectionStateError),
			LastError:       strPtr(err.Error()),
		})
		return err
	}

	// 构建 WebSocket URL
	url := fmt.Sprintf("%s?app_id=%s&token=%s&version=%s",
		c.account.WSEndpoint, appId, token, pluginVersion)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		c.isConnecting = false
		c.emitStatus(&RuntimeStatusPatch{
			Running:         boolPtr(true),
			Connected:       boolPtr(false),
			ConnectionState: ptrConnectionState(ConnectionStateError),
			LastError:       strPtr(err.Error()),
		})

		if c.shouldReconnect && c.options.AutoReconnect {
			c.scheduleReconnect(err.Error())
		}
		return err
	}

	c.ws = conn
	c.isConnecting = false
	c.reconnectAttempts = 0
	c.lastPongTime = time.Now().UnixMilli()

	c.emitStatus(&RuntimeStatusPatch{
		Running:           boolPtr(true),
		Connected:         boolPtr(true),
		ConnectionState:   ptrConnectionState(ConnectionStateConnected),
		ReconnectAttempts: intPtr(0),
		NextRetryAt:       nil,
		LastConnectedAt:   int64Ptr(time.Now().UnixMilli()),
		LastError:         nil,
	})

	c.startHeartbeat()
	c.options.OnOpen()

	// 启动读取 goroutine
	go c.readLoop()

	return nil
}

// readLoop 读取循环
func (c *WebSocketClient) readLoop() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.options.OnError(fmt.Errorf("WebSocket read error: %w", err))
			}

			c.mu.RLock()
			shouldReconnect := c.shouldReconnect
			autoReconnect := c.options.AutoReconnect
			code := websocket.CloseGoingAway
			reason := err.Error()
			c.mu.RUnlock()

			c.stopHeartbeat()

			if strings.Contains(reason, "4002") || strings.Contains(strings.ToLower(reason), "invalid token") {
				c.tokenManager.ClearCache()
			}

			c.emitStatus(&RuntimeStatusPatch{
				Running:         boolPtr(shouldReconnect),
				Connected:       boolPtr(false),
				ConnectionState: ptrConnectionState(ifThenElse(shouldReconnect && autoReconnect, ConnectionStateError, ConnectionStateStopped)),
				LastDisconnect: &DisconnectInfo{
					Code:   int(code),
					Reason: reason,
					At:     time.Now().UnixMilli(),
				},
			})

			c.options.OnClose(code, reason)

			if shouldReconnect && autoReconnect {
				c.scheduleReconnect(reason)
			}

			return
		}

		text := string(message)

		// 处理 pong 响应
		if text == "pong" || text == `{"type":"pong"}` {
			c.lastPongTime = time.Now().UnixMilli()
			continue
		}

		// 解析消息
		var data any
		if err := json.Unmarshal(message, &data); err != nil {
			continue
		}

		c.options.OnMessage(data)
	}
}

// startHeartbeat 启动心跳
func (c *WebSocketClient) startHeartbeat() {
	c.stopHeartbeat()

	c.pingInterval = time.AfterFunc(PingInterval, func() {
		c.sendHeartbeat()
	})
}

// sendHeartbeat 发送心跳
func (c *WebSocketClient) sendHeartbeat() {
	c.mu.RLock()
	ws := c.ws
	c.mu.RUnlock()

	if ws == nil {
		return
	}

	// 检查是否收到 pong
	if time.Since(time.UnixMilli(c.lastPongTime)) > PingInterval+c.pingTimeout {
		log.Printf("weibo[%s]: pong timeout, closing connection", c.account.AccountId)
		c.mu.Lock()
		c.ws = nil
		c.mu.Unlock()
		ws.Close()
		return
	}

	// 发送 ping
	if err := ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"ping"}`)); err != nil {
		log.Printf("weibo[%s]: failed to send ping: %v", c.account.AccountId, err)
	}

	// 调度下一次心跳
	c.pingInterval.Reset(PingInterval)
}

// stopHeartbeat 停止心跳
func (c *WebSocketClient) stopHeartbeat() {
	if c.pingInterval != nil {
		c.pingInterval.Stop()
		c.pingInterval = nil
	}
}

// scheduleReconnect 调度重连
func (c *WebSocketClient) scheduleReconnect(lastError string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.reconnectTimeout != nil {
		return
	}

	// 检查最大重连次数
	if c.options.MaxReconnectAttempts > 0 && c.reconnectAttempts >= c.options.MaxReconnectAttempts {
		log.Printf("weibo[%s]: max reconnect attempts reached", c.account.AccountId)
		return
	}

	// 计算延迟（指数退避）
	delay := min(InitialReconnectDelay*time.Duration(1<<c.reconnectAttempts), MaxReconnectDelay)

	c.reconnectAttempts++
	nextRetryAt := time.Now().Add(delay).UnixMilli()

	c.emitStatus(&RuntimeStatusPatch{
		Running:           boolPtr(true),
		Connected:         boolPtr(false),
		ConnectionState:   ptrConnectionState(ConnectionStateBackoff),
		ReconnectAttempts: intPtr(c.reconnectAttempts),
		NextRetryAt:       &nextRetryAt,
		LastError:         strPtr(lastError),
	})

	log.Printf("weibo[%s]: reconnecting in %v (attempt %d)", c.account.AccountId, delay, c.reconnectAttempts)

	c.reconnectTimeout = time.AfterFunc(delay, func() {
		c.mu.Lock()
		c.reconnectTimeout = nil
		c.mu.Unlock()

		if err := c.Connect(); err != nil {
			log.Printf("weibo[%s]: reconnect failed: %v", c.account.AccountId, err)
		}
	})
}

// emitStatus 发送状态更新
func (c *WebSocketClient) emitStatus(patch *RuntimeStatusPatch) {
	if c.options.OnStatus != nil {
		c.options.OnStatus(patch)
	}
}

// Send 发送消息
func (c *WebSocketClient) Send(data any) bool {
	c.mu.RLock()
	ws := c.ws
	c.mu.RUnlock()

	if ws == nil {
		return false
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("weibo[%s]: failed to marshal message: %v", c.account.AccountId, err)
		return false
	}

	if err := ws.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		log.Printf("weibo[%s]: failed to send message: %v", c.account.AccountId, err)
		return false
	}

	return true
}

// IsConnected 检查是否已连接
func (c *WebSocketClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ws != nil
}

// Close 关闭连接
func (c *WebSocketClient) Close() {
	c.mu.Lock()
	c.shouldReconnect = false
	c.mu.Unlock()

	c.stopHeartbeat()

	c.mu.Lock()
	if c.reconnectTimeout != nil {
		c.reconnectTimeout.Stop()
		c.reconnectTimeout = nil
	}
	c.mu.Unlock()

	c.mu.Lock()
	if c.ws != nil {
		c.ws.Close()
		c.ws = nil
	}
	c.mu.Unlock()

	now := time.Now().UnixMilli()
	c.emitStatus(&RuntimeStatusPatch{
		Running:         boolPtr(false),
		Connected:       boolPtr(false),
		ConnectionState: ptrConnectionState(ConnectionStateStopped),
		NextRetryAt:     nil,
		LastStopAt:      &now,
	})
}

// OnMessage 设置消息处理器
func (c *WebSocketClient) OnMessage(handler WebSocketMessageHandler) {
	c.options.OnMessage = handler
}

// OnError 设置错误处理器
func (c *WebSocketClient) OnError(handler WebSocketErrorHandler) {
	c.options.OnError = handler
}

// OnClose 设置关闭处理器
func (c *WebSocketClient) OnClose(handler WebSocketCloseHandler) {
	c.options.OnClose = handler
}

// OnOpen 设置打开处理器
func (c *WebSocketClient) OnOpen(handler WebSocketOpenHandler) {
	c.options.OnOpen = handler
}

// OnStatus 设置状态处理器
func (c *WebSocketClient) OnStatus(handler WebSocketStatusHandler) {
	c.options.OnStatus = handler
}
