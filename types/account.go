package types

// Account 账号信息
type Account struct {
	AccountID              string   // 账号ID
	Enabled                bool     // 是否启用
	Configured             bool     // 是否已配置
	AppID                  string   // 应用ID
	AppSecret              string   // 应用密钥
	TokenEndpoint          string   // Token 端点
	WSEndpoint             string   // WebSocket 端点
	WsMaxReconnectAttempts int      // WebSocket 最大重连尝试次数
	DMPolicy               string   // 私信策略
	AllowFrom              []string // 允许列表 允许发送私信的用户 ID 列表（白名单）
	TextChunkLimit         *int     // 文本分片限制 单条消息最大字符数，超出后自动分片
	ChunkMode              string   // 分片模式

	Config *WeiboConfig // 微博配置
}

// AccountListAccountsResponse 列出所有账号 响应
type AccountListAccountsResponse struct {
	Total    int        // 总数
	Accounts []*Account // 账号列表
}
