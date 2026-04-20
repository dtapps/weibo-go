package message

import (
	"encoding/json"
	"strings"

	"github.com/dtapps/weibo-go/types"
)

// ToInboundMessage WsMessageMsg 转换为 InboundMessage
func ToInboundMessage(m *types.WsMessageMsg) *types.InboundMessage {

	// 结构体转换为[]byte
	rawMessage, _ := json.Marshal(m.Payload)

	inbound := &types.InboundMessage{
		MessageID: m.Payload.MessageID, // 消息ID

		SenderID:  m.Payload.FromUserID, // 发送者ID
		Timestamp: m.Payload.TimeStamp,  // 发送时间戳

		RecipientID: m.Payload.ToUserID, // 接收者ID

		RawMessage: rawMessage, // 原始数据
	}

	// 处理消息内容
	for _, item := range m.Payload.Input {
		if item.Role != "user" {
			continue
		}
		for _, content := range item.Content {
			if content.Type == "input_text" {
				inbound.Content = append(inbound.Content, types.MessageSegment{
					Type: "text",       // 消息类型 text | image | file
					Text: content.Text, // 文本内容
				})
			}
			if content.Type == "input_image" {
				inbound.Content = append(inbound.Content, types.MessageSegment{
					Type:     "image",                                  // 消息类型 text | image | file
					Data:     content.Source.Data,                      // Base64 编码的内容
					FileName: content.Filename,                         // 文件名
					FileSize: getBase64ActualSize(content.Source.Data), // 文件大小
					MimeType: content.Source.MediaType,                 // MIME 类型
				})
			}
		}
	}

	return inbound
}

// getBase64ActualSize 获取 Base64 编码字符串的实际字节数
func getBase64ActualSize(base64Str string) int64 {
	if base64Str == "" {
		return 0
	}

	// 找到内容开始的位置（排除 data:image/png;base64, 等前缀）
	if idx := strings.Index(base64Str, ","); idx != -1 {
		base64Str = base64Str[idx+1:]
	}

	n := len(base64Str)
	if n == 0 {
		return 0
	}

	// 统计末尾的占位符 =
	padding := 0
	if strings.HasSuffix(base64Str, "==") {
		padding = 2
	} else if strings.HasSuffix(base64Str, "=") {
		padding = 1
	}

	// 计算原始字节数
	return int64(n*3/4 - padding)
}
