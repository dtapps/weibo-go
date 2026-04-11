package weibo

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

// processedMessages 用于消息去重的集合
var processedMessages = &messageDedup{
	messages: make(map[string]bool),
	maxSize:  1000,
}

// messageDedup 消息去重
type messageDedup struct {
	mu       sync.RWMutex
	messages map[string]bool
	maxSize  int
}

func (d *messageDedup) isDuplicate(id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.messages[id] {
		return true
	}

	d.messages[id] = true

	if len(d.messages) > d.maxSize {
		// 删除一半旧条目
		count := 0
		for k := range d.messages {
			delete(d.messages, k)
			count++
			if count >= d.maxSize/2 {
				break
			}
		}
	}

	return false
}

func (d *messageDedup) clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.messages = make(map[string]bool)
}

// NormalizedInboundInput 标准化的入站输入
type NormalizedInboundInput struct {
	Text   string
	Images []InboundAttachment
	Files  []InboundAttachment
}

// InboundAttachment 入站附件
type InboundAttachment struct {
	MimeType string
	Filename string
	Data     string
}

// MessageHandler 消息处理器
type MessageHandler struct {
	config  *WeiboConfig
	onReply func(to string, text string) error
	onLog   func(format string, args ...any)
	onError func(format string, args ...any)
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(config *WeiboConfig, onReply func(to string, text string) error) *MessageHandler {
	return &MessageHandler{
		config:  config,
		onReply: onReply,
		onLog:   func(format string, args ...any) { fmt.Printf(format+"\n", args...) },
		onError: func(format string, args ...any) { fmt.Printf("ERROR: "+format+"\n", args...) },
	}
}

// SetLogger 设置日志处理器
func (h *MessageHandler) SetLogger(logFunc func(format string, args ...any)) {
	h.onLog = logFunc
}

// SetErrorHandler 设置错误处理器
func (h *MessageHandler) SetErrorHandler(errFunc func(format string, args ...any)) {
	h.onError = errFunc
}

// HandleMessage 处理消息
func (h *MessageHandler) HandleMessage(event *InboundMessage) (*MessageContext, error) {
	payload := event.Payload

	// 获取消息 ID
	messageId := ResolveInboundMessageId(event)

	// 去重检查
	if processedMessages.isDuplicate(messageId) {
		return nil, nil
	}

	account := ResolveAccount(h.config, "")

	// 获取配置
	allowFrom := account.Config.AllowFrom
	dmPolicy := account.Config.DMPolicy

	// 检查 DM 策略
	if dmPolicy == "pairing" && len(allowFrom) > 0 {
		allowed := slices.Contains(allowFrom, payload.FromUserId)
		if !allowed {
			h.onLog("消息被拒绝: 用户 %s 不在白名单中", payload.FromUserId)
			return nil, nil
		}
	}

	// 规范化输入
	normalized := normalizeInboundInput(event)

	// 检查是否有内容
	hasText := len(strings.TrimSpace(normalized.Text)) > 0
	hasAttachments := len(normalized.Images) > 0 || len(normalized.Files) > 0

	if !hasText && !hasAttachments {
		return nil, nil
	}

	h.onLog("收到消息 from=%s, text=%s", payload.FromUserId, normalized.Text)

	// 返回消息上下文
	createTime := int64(0)
	if payload.Timestamp != nil {
		createTime = *payload.Timestamp
	}

	return &MessageContext{
		MessageId:  messageId,
		SenderId:   payload.FromUserId,
		Text:       normalized.Text,
		CreateTime: &createTime,
	}, nil
}

// normalizeInboundInput 规范化入站输入
func normalizeInboundInput(event *InboundMessage) *NormalizedInboundInput {
	result := &NormalizedInboundInput{}

	var textParts []string

	for _, item := range event.Payload.Input {
		if item.Type != "message" || item.Role != "user" {
			continue
		}

		for _, part := range item.Content {
			switch part.Type {
			case "input_text":
				if part.Text != nil && *part.Text != "" {
					textParts = append(textParts, *part.Text)
				}
			case "input_image":
				if part.Source != nil {
					result.Images = append(result.Images, InboundAttachment{
						MimeType: part.Source.MediaType,
						Filename: derefString(part.Filename, ""),
						Data:     part.Source.Data,
					})
				}
			case "input_file":
				if part.Source != nil {
					result.Files = append(result.Files, InboundAttachment{
						MimeType: part.Source.MediaType,
						Filename: derefString(part.Filename, ""),
						Data:     part.Source.Data,
					})
				}
			}
		}
	}

	if len(textParts) > 0 {
		result.Text = strings.Join(textParts, "\n")
	} else if event.Payload.Text != nil {
		result.Text = *event.Payload.Text
	}

	return result
}

// derefString 解引用字符串指针

// OutboundStream 出站流处理器
type OutboundStream struct {
	mu             sync.Mutex
	textChunkLimit int
	chunkMode      string
	emit           func(text string, done bool) error
	pendingText    string
}

// NewOutboundStream 创建出站流
func NewOutboundStream(chunkLimit int, chunkMode string, emit func(text string, done bool) error) *OutboundStream {
	if chunkLimit <= 0 {
		chunkLimit = 2000
	}
	return &OutboundStream{
		textChunkLimit: chunkLimit,
		chunkMode:      chunkMode,
		emit:           emit,
	}
}

// PushText 推送文本
func (s *OutboundStream) PushText(text string, done bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pendingText += text

	// 根据模式分片
	var chunks []string
	switch s.chunkMode {
	case "newline":
		chunks = s.splitByNewline(s.pendingText)
	case "length":
		chunks = s.splitByLength(s.pendingText)
	default:
		// raw 模式：直接发送
		chunks = []string{s.pendingText}
		s.pendingText = ""
	}

	// 发送所有完整的分片
	for i := 0; i < len(chunks)-1; i++ {
		if err := s.emit(chunks[i], false); err != nil {
			return err
		}
	}

	// 如果还有剩余文本且不是最后一块，保持在缓冲区
	if len(chunks) > 0 && !done {
		s.pendingText = chunks[len(chunks)-1]
	} else if len(chunks) > 0 {
		// 发送最后一块
		if err := s.emit(chunks[len(chunks)-1], done); err != nil {
			return err
		}
		s.pendingText = ""
	}

	return nil
}

// Flush 刷新缓冲区
func (s *OutboundStream) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pendingText != "" {
		if err := s.emit(s.pendingText, true); err != nil {
			return err
		}
		s.pendingText = ""
	}
	return nil
}

// splitByNewline 按换行符分割
func (s *OutboundStream) splitByNewline(text string) []string {
	var lines []string
	var current strings.Builder

	for _, line := range strings.Split(text, "\n") {
		if current.Len()+len(line) > s.textChunkLimit {
			if current.Len() > 0 {
				lines = append(lines, current.String())
				current.Reset()
			}
			// 如果单行本身超过限制，按字符数分割
			for len(line) > s.textChunkLimit {
				lines = append(lines, line[:s.textChunkLimit])
				line = line[s.textChunkLimit:]
			}
			if len(line) > 0 {
				current.WriteString(line)
			}
		} else {
			if current.Len() > 0 {
				current.WriteString("\n")
			}
			current.WriteString(line)
		}
	}

	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	return lines
}

// splitByLength 按字符数分割
func (s *OutboundStream) splitByLength(text string) []string {
	var chunks []string
	for i := 0; i < len(text); i += s.textChunkLimit {
		end := min(i+s.textChunkLimit, len(text))
		chunks = append(chunks, text[i:end])
	}
	return chunks
}
