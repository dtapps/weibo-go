package types

// ResolvedAccount 已解析的账号信息
type ResolvedAccount struct {
	AccountID     string        // 账号ID
	Name          string        // 账号名称
	Enabled       bool          // 是否启用
	Configured    bool          // 是否已配置
	AppId         string        // 应用ID
	AppSecret     string        // 应用密钥
	BotId         string        // 机器人ID
	TokenEndpoint string        // Token 端点
	WsGatewayUrl  string        // WebSocket 网关地址
	Config        AccountConfig // 账号配置
}

// AccountConfig 账号配置
type AccountConfig struct {
	DMPolicy       string   `json:"dmPolicy"`                 // 私信策略
	AllowFrom      []string `json:"allowFrom"`                // 允许接收的消息来源（预留）
	TokenEndpoint  string   `json:"tokenEndpoint"`            // Token 端点
	WSEndpoint     string   `json:"wsEndpoint"`               // WebSocket 端点
	TextChunkLimit *int     `json:"textChunkLimit,omitempty"` // 文本分块限制（预留）
	ChunkMode      string   `json:"chunkMode"`                // 分块模式
	BlockStreaming *bool    `json:"blockStreaming,omitempty"` // 阻止流式处理（预留）
}

// TokenResult Token 结果
type TokenResult struct {
	Token      string `json:"token"`      // Token
	ExpiresIn  int    `json:"expiresIn"`  // 过期时间（秒）
	AcquiredAt int64  `json:"acquiredAt"` // 获取时间（毫秒时间戳）
	Uid        int64  `json:"uid"`        // 用户ID
}

// TokenResponse Token 响应
type TokenResponse struct {
	Data TokenData `json:"data"`
}

// TokenData Token 数据
type TokenData struct {
	Token    string `json:"token"`     // Token
	ExpireIn int    `json:"expire_in"` // 过期时间（秒）
	Uid      int64  `json:"uid"`       // 用户ID
}

// TokenCache Token 缓存
type TokenCache struct {
	Token      string // Token
	AcquiredAt int64  // 获取时间（毫秒时间戳）
	ExpiresIn  int    // 过期时间（秒）
}
