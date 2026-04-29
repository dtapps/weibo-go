package types

// TokenCallback Token 回调数据
type TokenCallbackData struct {
	Status     string // 状态：success | error
	AppID      string // 应用 ID
	Token      string // Token
	ExpiresIn  int64  // 过期时长（秒）
	AcquiredAt int64  // 获取时间（秒）
	ExpiresAt  int64  // 过期时间（秒）
	Error      error  // 错误信息
}

// TokenCallback Token 回调函数
type TokenCallback func(data *TokenCallbackData)

// TokenCache Token 缓存
type TokenCache struct {
	Token      string // Token
	AcquiredAt int64  // 获取时间（秒）
	ExpiresAt  int64  // 过期时间（秒）
}

// TokenResult Token 结果
type TokenResult struct {
	AppID      int64  // 应用 ID
	Token      string // Token
	ExpiresIn  int64  // 过期时长（秒）
	AcquiredAt int64  // 获取时间（秒）
}

// TokenRequest Token 请求
type TokenRequest struct {
	AppID     string `json:"app_id"`     // 应用ID
	AppSecret string `json:"app_secret"` // 应用密钥
}

// TokenResponse Token 响应
type TokenResponse struct {
	Code    int       `json:"code,omitempty"`    // 状态码
	Message string    `json:"message,omitempty"` // 状态描述
	Data    TokenData `json:"data"`              // Token 数据
}

// TokenData Token 数据
type TokenData struct {
	Uid      int64  `json:"uid,omitempty"`       // 用户ID
	ExpireIn int64  `json:"expire_in,omitempty"` // 过期时长（秒）
	Token    string `json:"token,omitempty"`     // Token
}
