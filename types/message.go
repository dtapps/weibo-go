package types

// SendResult 发送结果
type SendResult struct {
	Ok        bool
	MessageID string
	Error     error
}
