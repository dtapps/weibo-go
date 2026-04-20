package types

import "time"

const (
	// 默认的 API 端点
	DefaultWSEndpoint        = "ws://open-im.api.weibo.com/ws/stream"
	DefaultTokenEndpoint     = "https://open-im.api.weibo.com/open/auth/ws_token"
	DefaultHotSearchEndpoint = "https://open-im.api.weibo.com/open/weibo/hot_search"
	DefaultSearchEndpoint    = "https://open-im.api.weibo.com/open/wis/search_query"
	DefaultStatusEndpoint    = "https://open-im.api.weibo.com/open/weibo/user_status"

	// TokenExpireDuration token 过期时间（秒）
	TokenExpireDuration = 7200 * time.Second
	// TokenRefreshBuffer token 刷新缓冲时间
	TokenRefreshBuffer = 60 * time.Second

	// TokenFetchBaseDelay token 获取基础延迟
	TokenFetchBaseDelay = 1 * time.Second
	// TokenFetchMaxDelay token 获取最大延迟
	TokenFetchMaxDelay = 8 * time.Second
	// TokenFetchMaxRetries token 获取最大重试次数
	TokenFetchMaxRetries = 2

	// TokenRefreshCheckInterval Token 刷新检查间隔
	TokenRefreshCheckInterval = 30 * time.Second
	// TokenRefreshThreshold Token 刷新阈值（提前多久刷新）
	TokenRefreshThreshold = 5 * time.Minute

	// 默认心跳间隔(秒)
	DefaultHeartbeatInterval = 30 * time.Second
	// 默认心跳超时次数阈值
	HeartbeatTimeoutThreshold = 2
	// 默认最大重连次数
	DefaultMaxReconnectAttempts = 100
	// 默认重连延迟
	DefaultReconnectDelays = "1s,2s,5s,10s,30s,60s"
)

// 榜单类型映射：中文名称 -> 内部标识
var CategoryMap = map[string]string{
	"主榜":   "v_openclaw",
	"文娱榜":  "v_openclaw_ent",
	"社会榜":  "v_openclaw_social",
	"生活榜":  "v_openclaw_live",
	"acg榜": "v_openclaw_acg",
	"科技榜":  "v_openclaw_tech",
	"体育榜":  "v_openclaw_sport",
}

// 可重试的 HTTP 状态码
var RetryableStatusCodes = map[int]bool{
	408: true, 425: true, 429: true,
	500: true, 502: true, 503: true, 504: true,
}

// ConnectionState 连接状态
type ConnectionState string

const (
	ConnectionStateIdle         ConnectionState = "idle"         // 空闲
	ConnectionStateConnecting   ConnectionState = "connecting"   // 连接中
	ConnectionStateConnected    ConnectionState = "connected"    // 已连接
	ConnectionStateBackoff      ConnectionState = "backoff"      // 重试中
	ConnectionStateReconnecting ConnectionState = "reconnecting" // 重连中
	ConnectionStateError        ConnectionState = "error"        // 错误
	ConnectionStateStopped      ConnectionState = "stopped"      // 已停止
	ConnectionStateDisconnected ConnectionState = "disconnected" // 已断开
)

// String 连接状态 字符串
func (c ConnectionState) String() string {
	return string(c)
}
