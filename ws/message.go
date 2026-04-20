package ws

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/types"
	"github.com/gorilla/websocket"
)

// SendMessage 发送消息
func (c *WsClient) SendMessage(toUserID string, text string, messageID string, chunkID int, done bool) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("连接未建立")
	}

	msg := types.SendMessageRequest{
		Type: "send_message",
		Payload: types.SendMessagePayload{
			ToUserID:  toUserID,
			Text:      strings.TrimSpace(text),
			MessageID: messageID,
			ChunkID:   chunkID,
			Done:      done,
		},
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("构建消息失败: %w", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	c.log.Debug("发送消息成功",
		logger.F("messageID", messageID),
		logger.F("data", string(data)),
	)

	return nil
}
