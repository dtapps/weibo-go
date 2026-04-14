package weibo

import (
	"github.com/dtapps/weibo-go/account"
	"github.com/dtapps/weibo-go/member"
	"github.com/dtapps/weibo-go/plugin"
	"github.com/dtapps/weibo-go/types"
)

type Client struct {
	plugin *plugin.Plugin
}

func NewClient(accountId string, cfg *types.Config) (*Client, error) {
	acc := account.GetManager().ResolveAccount(cfg, accountId)

	if !acc.Configured {
		return nil, ErrAccountNotConfigured
	}

	p, err := plugin.CreateAndStart(accountId, acc, cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		plugin: p,
	}, nil
}

// OnMessage 设置消息处理回调
func (c *Client) OnMessage(handler func(msg *types.WsMessageMsg)) {
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

// SendMessage 发送消息
func (c *Client) SendMessage(to string, text string) error {
	_, err := c.plugin.SendMessage(to, text)
	return err
}

// SendMessageChunked 分片发送消息
func (c *Client) SendMessageChunked(to string, text string, chunkLimit int) ([]*types.SendResult, error) {
	if chunkLimit <= 0 {
		chunkLimit = 2000
	}

	runes := []rune(text)
	length := len(runes)

	if length <= chunkLimit {
		result, err := c.plugin.SendMessage(to, text)
		if err != nil {
			return nil, err
		}
		return []*types.SendResult{result}, nil
	}

	var results []*types.SendResult
	for i := 0; i < length; i += chunkLimit {
		end := i + chunkLimit
		if end > length {
			end = length
		}
		chunk := string(runes[i:end])
		result, err := c.plugin.SendMessage(to, chunk)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}

	return results, nil
}

func (c *Client) GetState() string {
	return c.plugin.GetState()
}

func (c *Client) GetMember() member.MemberInterface {
	return c.plugin.GetMember()
}

func (c *Client) GetTokenManager() *account.TokenManager {
	return c.plugin.GetTokenManager()
}

func (c *Client) Stop() error {
	return c.plugin.Stop()
}

type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

var (
	ErrAccountNotConfigured = &Error{Code: 1, Message: "账号未配置"}
	ErrNotConnected         = &Error{Code: 2, Message: "未连接到服务器"}
	ErrSendFailed           = &Error{Code: 3, Message: "发送消息失败"}
)
