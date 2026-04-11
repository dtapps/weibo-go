package weibo

import (
	"encoding/json"
	"fmt"
)

// Client 微博客户端
type Client struct {
	config        *WeiboConfig
	clientManager *ClientManager
}

// NewClient 创建新的微博客户端
func NewClient(config *WeiboConfig) *Client {
	return &Client{
		config:        config,
		clientManager: NewClientManager(),
	}
}

// NewClientWithAppCredentials 使用 App 凭证创建客户端
func NewClientWithAppCredentials(appId, appSecret string) *Client {
	config := &WeiboConfig{
		Enabled:       boolPtr(true),
		AppId:         &appId,
		AppSecret:     &appSecret,
		WSEndpoint:    strPtr(DefaultWSEndpoint),
		TokenEndpoint: strPtr(DefaultTokenEndpoint),
		DMPolicy:      strPtr("open"),
		ChunkMode:     strPtr("raw"),
	}
	return NewClient(config)
}

// Connect 连接到微博 WebSocket 服务
func (c *Client) Connect(opts *ConnectOptions) error {
	if opts == nil {
		opts = &ConnectOptions{}
	}

	options := &MonitorOptions{
		AccountId:     opts.AccountId,
		AutoReconnect: true,
		OnMessage: func(msg *InboundMessage) {
			if opts.OnMessage != nil {
				opts.OnMessage(msg)
			}
		},
		OnError: func(err error) {
			if opts.OnError != nil {
				opts.OnError(err)
			}
		},
		OnClose: func(code int, reason string) {
			if opts.OnClose != nil {
				opts.OnClose(code, reason)
			}
		},
		OnOpen: func() {
			if opts.OnOpen != nil {
				opts.OnOpen()
			}
		},
		OnStatus: opts.OnStatus,
	}

	return Monitor(c.clientManager, c.config, options)
}

// ConnectOptions 连接选项
type ConnectOptions struct {
	AccountId *string
	OnMessage func(*InboundMessage)
	OnError   func(error)
	OnClose   func(code int, reason string)
	OnOpen    func()
	OnStatus  func(*RuntimeStatusPatch)
}

// Send 发送消息
func (c *Client) Send(to, text string) (*SendResult, error) {
	return SendMessage(c.clientManager, c.config, &SendMessageParams{
		To:   to,
		Text: text,
		Done: true,
	})
}

// SendChunked 分片发送消息
func (c *Client) SendChunked(to, text string, chunkLimit int) ([]*SendResult, error) {
	if chunkLimit <= 0 {
		chunkLimit = 2000
	}

	var results []*SendResult

	// 按字符数分片
	for i := 0; i < len(text); i += chunkLimit {
		end := min(i+chunkLimit, len(text))

		chunk := text[i:end]
		done := end >= len(text)

		result, err := SendMessage(c.clientManager, c.config, &SendMessageParams{
			To:      to,
			Text:    chunk,
			ChunkId: len(results),
			Done:    done,
		})

		if err != nil {
			return results, err
		}

		results = append(results, result)
	}

	return results, nil
}

// GetHotSearch 获取热搜
func (c *Client) GetHotSearch(category string, count *int) (*HotSearchResult, error) {
	account := ResolveAccount(c.config, "")
	if !account.Configured {
		return nil, fmt.Errorf("微博账号未配置")
	}

	tokenManager := c.clientManager.GetTokenManager()
	token, err := tokenManager.GetValidToken(
		*account.AppId,
		*account.AppSecret,
		account.TokenEndpoint,
	)
	if err != nil {
		return nil, fmt.Errorf("获取 token 失败: %w", err)
	}

	return FetchHotSearch(token, category, count, "")
}

// Search 搜索微博
func (c *Client) Search(query string) (*SearchResult, error) {
	account := ResolveAccount(c.config, "")
	if !account.Configured {
		return nil, fmt.Errorf("微博账号未配置")
	}

	tokenManager := c.clientManager.GetTokenManager()
	token, err := tokenManager.GetValidToken(
		*account.AppId,
		*account.AppSecret,
		account.TokenEndpoint,
	)
	if err != nil {
		return nil, fmt.Errorf("获取 token 失败: %w", err)
	}

	return SearchWeibo(query, token, "")
}

// GetStatus 获取用户微博
func (c *Client) GetStatus(count *int) (*StatusResult, error) {
	account := ResolveAccount(c.config, "")
	if !account.Configured {
		return nil, fmt.Errorf("微博账号未配置")
	}

	tokenManager := c.clientManager.GetTokenManager()
	token, err := tokenManager.GetValidToken(
		*account.AppId,
		*account.AppSecret,
		account.TokenEndpoint,
	)
	if err != nil {
		return nil, fmt.Errorf("获取 token 失败: %w", err)
	}

	return FetchWeiboStatus(token, count, "")
}

// GetToken 获取当前 token
func (c *Client) GetToken() (string, error) {
	account := ResolveAccount(c.config, "")
	if !account.Configured {
		return "", fmt.Errorf("微博账号未配置")
	}

	tokenManager := c.clientManager.GetTokenManager()
	return tokenManager.GetValidToken(
		*account.AppId,
		*account.AppSecret,
		account.TokenEndpoint,
	)
}

// Close 关闭客户端
func (c *Client) Close() {
	c.clientManager.ClearAllClients()
}

// GetConfig 获取配置
func (c *Client) GetConfig() *WeiboConfig {
	return c.config
}

// SetLogger 设置日志处理器
func (c *Client) SetLogger(logFunc func(format string, args ...any)) {
	// 可以在这里添加全局日志设置
	_ = logFunc
}

// ConfigFromJSON 从 JSON 创建配置
func ConfigFromJSON(jsonStr string) (*WeiboConfig, error) {
	var config WeiboConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// ConfigToJSON 将配置转换为 JSON
func ConfigToJSON(config *WeiboConfig) (string, error) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
