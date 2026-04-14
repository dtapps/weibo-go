package types

// WeiboConfig 微博配置结构
type WeiboConfig struct {
	Enabled        *bool          `json:"enabled,omitempty"`
	AppId          string         `json:"appId,omitempty"`
	AppSecret      string         `json:"appSecret,omitempty"`
	WSEndpoint     string         `json:"wsEndpoint,omitempty"`
	TokenEndpoint  string         `json:"tokenEndpoint,omitempty"`
	DMPolicy       string         `json:"dmPolicy,omitempty"`
	AllowFrom      []any          `json:"allowFrom,omitempty"`
	TextChunkLimit *int           `json:"textChunkLimit,omitempty"`
	ChunkMode      string         `json:"chunkMode,omitempty"`
	BlockStreaming *bool          `json:"blockStreaming,omitempty"`
	Accounts       map[string]any `json:"accounts,omitempty"`
}

// Config 全局配置
type Config struct {
	Weibo *WeiboConfig `json:"weibo,omitempty"`
}
