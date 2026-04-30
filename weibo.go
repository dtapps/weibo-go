package weibo

import (
	"fmt"

	"github.com/dtapps/weibo-go/account"
	"github.com/dtapps/weibo-go/member"
	"github.com/dtapps/weibo-go/plugin"
	"github.com/dtapps/weibo-go/token"
	"github.com/dtapps/weibo-go/types"
)

// 错误定义
var (
	ErrAccountNotConfigured = fmt.Errorf("账号未配置")
)

// WithTokenCallback 设置 Token 回调
func WithTokenCallback(handler func(data *types.TokenCallbackData)) plugin.ClientOption {
	return func(p *plugin.Plugin) {
		p.GetTokenManager().SetCallback(handler)
	}
}

// Client 微博客户端
type Client struct {
	plugin *plugin.Plugin
}

// NewClient 创建新客户端
func NewClient(accountID string, cfg *types.Config, options ...plugin.ClientOption) (*Client, error) {
	// 创建账号管理器
	accountMgr := account.NewManager()

	// 解析账号配置
	acc := accountMgr.ResolveAccount(cfg, accountID)
	if !acc.Configured {
		return nil, ErrAccountNotConfigured
	}

	// 创建插件管理器
	pluginMgr := plugin.NewPluginManager()

	// 创建并启动插件
	p, err := plugin.CreateAndStart(pluginMgr, accountID, acc, cfg, options...)
	if err != nil {
		return nil, err
	}

	return &Client{
		plugin: p,
	}, nil
}

// OnMessage 设置消息处理回调
func (c *Client) OnMessage(handler func(msg *types.InboundMessage)) {
	c.plugin.SetOnMessage(handler)
}

// OnConnected 设置连接成功回调
func (c *Client) OnConnected(handler func()) {
	c.plugin.SetOnConnected(handler)
}

// OnDisconnected 设置断开连接回调
func (c *Client) OnDisconnected(handler func()) {
	c.plugin.SetOnDisconnected(handler)
}

// OnError 设置错误回调
func (c *Client) OnError(handler func(err error)) {
	c.plugin.SetOnError(handler)
}

// SendMessage 发送消息
func (c *Client) SendMessage(msg *types.OutboundMessage) (string, error) {
	messageID, err := c.plugin.SendMessage(msg)
	if err != nil {
		return "", err
	}
	return messageID, nil
}

// SendMessageChunked 分片发送消息
func (c *Client) SendMessageChunked(msg *types.OutboundMessage, chunkLimit int) ([]string, error) {
	if chunkLimit <= 0 {
		chunkLimit = 2000
	}

	runes := []rune(msg.Text)
	length := len(runes)

	if length <= chunkLimit {
		result, err := c.plugin.SendMessage(msg)
		if err != nil {
			return nil, err
		}
		return []string{result}, nil
	}

	var results []string
	for i := 0; i < length; i += chunkLimit {
		end := min(i+chunkLimit, length)
		chunk := string(runes[i:end])
		result, err := c.plugin.SendMessage(&types.OutboundMessage{
			ToUserID: msg.ToUserID,
			Text:     chunk,
		})
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}

	return results, nil
}

// GetState 获取连接状态
func (c *Client) GetState() string {
	return c.plugin.GetState()
}

// GetMember 获取成员管理器
func (c *Client) GetMember() *member.Manager {
	return c.plugin.GetMember()
}

// GetTokenManager 获取 Token 管理器
func (c *Client) GetTokenManager() *token.Manager {
	return c.plugin.GetTokenManager()
}

// Stop 停止客户端
func (c *Client) Stop() error {
	return c.plugin.Stop()
}
