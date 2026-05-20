package ws

import (
	"encoding/json"
	"time"

	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/types"
)

// handleClose 处理关闭
func (c *WsClient) handleClose(code int, reason string) {
	c.log.Info("处理关闭",
		logger.F("code", code),
		logger.F("reason", reason),
	)

	c.mu.Lock()
	wasConnected := c.state == types.ConnectionStateConnected.String()
	isManualDisconnect := c.manualDisconnect
	// 重置手动断开标志
	c.manualDisconnect = false
	c.mu.Unlock()

	c.stopHeartbeat()

	c.close()

	c.mu.Lock()
	c.state = types.ConnectionStateDisconnected.String()
	c.mu.Unlock()

	if c.callback != nil {
		c.callback.OnClose(code, reason)
	}

	// 安排重连
	// 优先判断：手动断开不需要重连
	if isManualDisconnect {
		return
	}

	// 正常关闭 (code=1000) 不需要重连
	if code == 1000 {
		return
	}

	// 如果之前已连接或者是异常关闭，安排重连
	if wasConnected || code != 1000 {
		c.ScheduleReconnect()
	}
}

// handleMessage 处理消息
func (c *WsClient) handleMessage(data []byte) {
	c.mu.Lock()
	c.heartbeatAckReceived = true
	c.heartbeatTimeoutCount = 0
	c.mu.Unlock()

	text := string(data)

	// 处理 pong 类型
	if text == "pong" || text == `{"type":"pong"}` {
		c.lastHeartbeatAt = time.Now().UnixMilli()
		return
	}

	var msg types.WsMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		c.log.Error("解析消息失败", logger.F("error", err.Error()))
		return
	}

	switch msg.Type {
	case "connected":
		// 连接成功
		c.handleTypeConnected(data)
	case "message":
		// 消息
		c.handleTypeMessage(data)
	default:
		c.log.Warn("未处理的msgType", logger.F("msgType", msg.Type))
	}
}

// handleTypeConnected 处理连接成功类型
func (c *WsClient) handleTypeConnected(data []byte) {

	var msg types.WsConnectedMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		c.log.Error("解析消息失败", logger.F("error", err.Error()))
		return
	}

	c.mu.Lock()
	c.connectID = msg.ConnectionID
	c.state = types.ConnectionStateConnected.String()
	c.mu.Unlock()

	// 启动心跳定时器
	c.startHeartbeat()

	// 回调
	if c.callback != nil {
		result := &types.OnReadyData{
			ConnectID: c.connectID,
			Timestamp: time.Now().Unix(),
		}
		c.callback.OnReady(result)
		c.callback.OnStateChange(types.ConnectionStateConnected.String())
	}
}

// handleTypeMessage 处理消息类型
func (c *WsClient) handleTypeMessage(data []byte) {

	var msg types.WsMessageMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		c.log.Error("解析消息失败", logger.F("error", err.Error()))
		return
	}

	if c.callback != nil {
		c.callback.OnDispatch(&msg)
	}
}
