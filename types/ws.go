package types

// WsMsg WebSocket 消息
type WsMsg struct {
	Type string `json:"type"` // 消息类型
}

// WsConnectedMsg 连接成功消息
type WsConnectedMsg struct {
	WsMsg
	AppID        string `json:"app_id,omitempty"`        // 应用ID
	ConnectionID string `json:"connection_id,omitempty"` // 连接ID
	Message      string `json:"message,omitempty"`       // 消息内容
}

// WsMessageMsg 消息消息
type WsMessageMsg struct {
	WsMsg
	Payload struct {
		Input      []WsMessageMsgInputItem `json:"input,omitempty"`      // 输入参数
		FromUserId string                  `json:"fromUserId,omitempty"` // 发送用户ID
		MessageId  string                  `json:"messageId,omitempty"`  // 消息ID
		Text       string                  `json:"text,omitempty"`       // 消息内容
		ToUserId   string                  `json:"toUserId,omitempty"`   // 接收用户ID
		TimeStamp  int64                   `json:"timeStamp,omitempty"`  // 时间戳
	} `json:"payload"`
}

type WsMessageMsgInputItem struct {
	Role    string `json:"role,omitempty"` // 角色
	Type    string `json:"type,omitempty"` // 类型
	Content []struct {
		Type string `json:"type,omitempty"` // 类型 input_text | input_image
		// 文本
		Text string `json:"text,omitempty"` // 文本
		// 文件
		Filename string `json:"filename,omitempty"` // 文件名
		Source   struct {
			Data      string `json:"data,omitempty"`       // 数据
			MediaType string `json:"media_type,omitempty"` // 媒体类型
		} `json:"source"` // 来源
	} `json:"content,omitempty"` // 内容
}

// HeartbeatMsgRequest 发送心跳消息
type HeartbeatMsgRequest struct {
	Type string `json:"type"` // 消息类型
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Type    string `json:"type"` // 消息类型
	Payload struct {
		ToUserId  string `json:"toUserId,omitempty"`  // 接收用户ID
		Text      string `json:"text,omitempty"`      // 消息内容
		MessageId string `json:"messageId,omitempty"` // 消息ID
		ChunkId   int    `json:"chunkId,omitempty"`   // 分块ID
		Done      bool   `json:"done,omitempty"`      // 是否完成
	} `json:"payload"`
}

// SendMessagePayload 发送消息请求负载
type SendMessagePayload struct {
	ToUserId  string `json:"toUserId,omitempty"`  // 接收用户ID
	Text      string `json:"text,omitempty"`      // 消息内容
	MessageId string `json:"messageId,omitempty"` // 消息ID
	ChunkId   int    `json:"chunkId,omitempty"`   // 分块ID
	Done      bool   `json:"done,omitempty"`      // 是否完成
}
