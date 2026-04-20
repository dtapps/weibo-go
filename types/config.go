package types

// WeiboConfig 微博配置结构
type WeiboConfig struct {
	Enabled        *bool               `json:"enabled,omitempty"`        // 是否启用
	AppID          string              `json:"appID,omitempty"`          // 应用ID
	AppSecret      string              `json:"appSecret,omitempty"`      // 应用密钥
	TokenEndpoint  string              `json:"tokenEndpoint,omitempty"`  // Token 端点
	WSEndpoint     string              `json:"wsEndpoint,omitempty"`     // WebSocket 端点
	DMPolicy       string              `json:"dmPolicy,omitempty"`       // 私信策略
	AllowFrom      []any               `json:"allowFrom,omitempty"`      // 允许列表
	TextChunkLimit *int                `json:"textChunkLimit,omitempty"` // 文本分片限制
	ChunkMode      string              `json:"chunkMode,omitempty"`      // 分片模式
	GroupTrigger   *GroupTriggerConfig `json:"groupTrigger,omitempty"`   // 群聊触发配置（微博没有群聊功能）
}

// GroupTriggerConfig 群聊触发配置
type GroupTriggerConfig struct {
	MentionOnly bool     `json:"mention_only,omitempty"` // 是否仅响应@机器人
	Prefixes    []string `json:"prefixes,omitempty"`     // 前缀
}

// Config 全局配置
type Config struct {
	Weibo *WeiboConfig `json:"weibo,omitempty"`
}
