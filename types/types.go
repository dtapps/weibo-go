package types

// OutboundMessage 发送消息
type OutboundMessage struct {
	ToUserID string `json:"to_user_id,omitempty"` // 接收用户ID
	Text     string `json:"text,omitempty"`       // 消息内容
}

// InboundMessage 收到的消息
type InboundMessage struct {
	AccountID string `json:"account_id,omitempty"` // 账号ID
	AppID     string `json:"app_id,omitempty"`     // 应用ID
	MessageID string `json:"message_id,omitempty"` // 消息ID

	SenderID  string `json:"sender_id,omitempty"` // 发送者ID
	Timestamp int64  `json:"timestamp,omitempty"` // 发送时间戳

	RecipientID string `json:"recipient_id,omitempty"` // 接收者ID

	Content []MessageSegment `json:"content,omitempty"` // 消息内容

	RawMessage []byte `json:"-"` // 原始数据
}

// 消息内容
type MessageSegment struct {
	Type     string `json:"type"`                // 消息类型 text | image | file
	Text     string `json:"text,omitempty"`      // 文本内容
	Url      string `json:"url,omitempty"`       // 远程资源链接
	Data     string `json:"data,omitempty"`      // Base64 编码的内容
	FileName string `json:"file_name,omitempty"` // 文件名
	FileSize int64  `json:"file_size,omitempty"` // 文件大小
	MimeType string `json:"mime_type,omitempty"` // MIME 类型
}
