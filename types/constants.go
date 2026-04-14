package types

import "time"

const (
	// 默认的 API 端点
	DefaultWSEndpoint        = "ws://open-im.api.weibo.com/ws/stream"
	DefaultTokenEndpoint     = "http://open-im.api.weibo.com/open/auth/ws_token"
	DefaultHotSearchEndpoint = "http://open-im.api.weibo.com/open/weibo/hot_search"
	DefaultSearchEndpoint    = "http://open-im.api.weibo.com/open/wis/search_query"
	DefaultStatusEndpoint    = "http://open-im.api.weibo.com/open/weibo/user_status"

	// PingInterval 心跳间隔
	PingInterval = 30 * time.Second
	// PingTimeout 心跳超时时间
	PingTimeout = 10 * time.Second

	// InitialReconnectDelay 初始重连延迟
	InitialReconnectDelay = 1 * time.Second
	// MaxReconnectDelay 最大重连延迟
	MaxReconnectDelay = 60 * time.Second

	// TokenExpireSeconds token 过期时间（秒）
	TokenExpireSeconds = 7200
	// TokenRefreshBufferSeconds token 刷新缓冲时间（秒）
	TokenRefreshBufferSeconds = 60

	// TokenFetchBaseDelay token 获取基础延迟
	TokenFetchBaseDelay = 1 * time.Second
	// TokenFetchMaxDelay token 获取最大延迟
	TokenFetchMaxDelay = 8 * time.Second
	// TokenFetchMaxRetries token 获取最大重试次数
	TokenFetchMaxRetries = 2
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
