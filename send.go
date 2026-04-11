package weibo

import (
	"fmt"
)

// SendMessageParams 发送消息参数
type SendMessageParams struct {
	AccountId *string
	To        string
	Text      string
	MessageId string
	ChunkId   int
	Done      bool
}

// SendMessage 发送消息
func SendMessage(clientManager *ClientManager, config *WeiboConfig, params *SendMessageParams) (*SendResult, error) {
	account := ResolveAccount(config, "")
	if params.AccountId != nil && *params.AccountId != "" {
		account = ResolveAccount(config, *params.AccountId)
	}

	if !account.Configured {
		return nil, fmt.Errorf("微博账号 %s 未配置", account.AccountId)
	}

	client := clientManager.CreateClient(account, nil)

	receiveId := NormalizeWeiboTarget(params.To)
	if receiveId == "" {
		return nil, fmt.Errorf("无效的微博目标: %s", params.To)
	}

	userId := receiveId
	if len(userId) > 5 && userId[:5] == "user:" {
		userId = userId[5:]
	}

	messageId := params.MessageId
	if messageId == "" {
		messageId = GenerateMessageId()
	}

	chunkId := max(params.ChunkId, 0)

	done := params.Done
	if !done && params.Text != "" {
		done = true
	}

	payload := SendMessagePayload{
		ToUserId:  userId,
		Text:      params.Text,
		MessageId: messageId,
		ChunkId:   chunkId,
		Done:      done,
	}

	msg := WebSocketMessage{
		Type:    "send_message",
		Payload: payload,
	}

	if !client.Send(msg) {
		return nil, fmt.Errorf("发送消息失败: 客户端未连接")
	}

	return &SendResult{
		MessageId: messageId,
		ChatId:    receiveId,
		ChunkId:   chunkId,
		Done:      done,
	}, nil
}

// MonitorOptions 监控选项
type MonitorOptions struct {
	AccountId     *string
	OnMessage     func(msg *InboundMessage)
	OnError       func(err error)
	OnClose       func(code int, reason string)
	OnOpen        func()
	OnStatus      func(patch *RuntimeStatusPatch)
	AutoReconnect bool
}

// Monitor 启动监控
func Monitor(clientManager *ClientManager, config *WeiboConfig, opts *MonitorOptions) error {
	if opts == nil {
		opts = &MonitorOptions{
			AutoReconnect: true,
		}
	}

	var accounts []*ResolvedAccount
	if opts.AccountId != nil && *opts.AccountId != "" {
		account := ResolveAccount(config, *opts.AccountId)
		if !account.Enabled || !account.Configured {
			return fmt.Errorf("微博账号 %s 未配置或已禁用", *opts.AccountId)
		}
		accounts = []*ResolvedAccount{account}
	} else {
		accounts = ListEnabledAccounts(config)
		if len(accounts) == 0 {
			return fmt.Errorf("没有已配置的微博账号")
		}
	}

	for _, account := range accounts {
		options := &WebSocketClientOptions{
			OnMessage: func(data any) {
				handleWebSocketMessage(data, opts.OnMessage)
			},
			OnError:       opts.OnError,
			OnClose:       opts.OnClose,
			OnOpen:        opts.OnOpen,
			OnStatus:      opts.OnStatus,
			AutoReconnect: opts.AutoReconnect,
		}

		client := clientManager.CreateClient(account, options)
		if err := client.Connect(); err != nil {
			return err
		}
	}

	return nil
}

// handleWebSocketMessage 处理 WebSocket 消息
func handleWebSocketMessage(data any, handler func(*InboundMessage)) {
	if handler == nil {
		return
	}

	// 解析消息类型
	msgMap, ok := data.(map[string]any)
	if !ok {
		return
	}

	msgType, ok := msgMap["type"].(string)
	if !ok {
		return
	}

	if msgType == "message" {
		payload, ok := msgMap["payload"].(map[string]any)
		if !ok {
			return
		}

		event := &InboundMessage{
			Type:    msgType,
			Payload: parseInboundPayload(payload),
		}

		handler(event)
	}
}

// parseInboundPayload 解析入站载荷
func parseInboundPayload(payload map[string]any) InboundMessagePayload {
	result := InboundMessagePayload{}

	if v, ok := payload["messageId"].(string); ok {
		result.MessageId = v
	}

	if v, ok := payload["fromUserId"].(string); ok {
		result.FromUserId = v
	}

	if v, ok := payload["text"].(string); ok {
		result.Text = &v
	}

	if v, ok := payload["timestamp"].(float64); ok {
		timestamp := int64(v)
		result.Timestamp = &timestamp
	}

	if arr, ok := payload["input"].([]any); ok {
		for _, item := range arr {
			if itemMap, ok := item.(map[string]any); ok {
				inputItem := InboundMessageInputItem{}

				if t, ok := itemMap["type"].(string); ok {
					inputItem.Type = t
				}
				if r, ok := itemMap["role"].(string); ok {
					inputItem.Role = r
				}

				if content, ok := itemMap["content"].([]any); ok {
					for _, part := range content {
						if partMap, ok := part.(map[string]any); ok {
							cp := InboundContentPart{}

							if t, ok := partMap["type"].(string); ok {
								cp.Type = t
							}
							if t, ok := partMap["text"].(string); ok {
								cp.Text = &t
							}
							if f, ok := partMap["filename"].(string); ok {
								cp.Filename = &f
							}

							if source, ok := partMap["source"].(map[string]any); ok {
								cs := InboundContentSource{}
								if t, ok := source["type"].(string); ok {
									cs.Type = t
								}
								if mt, ok := source["media_type"].(string); ok {
									cs.MediaType = mt
								}
								if d, ok := source["data"].(string); ok {
									cs.Data = d
								}
								cp.Source = &cs
							}

							inputItem.Content = append(inputItem.Content, cp)
						}
					}
				}

				result.Input = append(result.Input, inputItem)
			}
		}
	}

	return result
}
