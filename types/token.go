package types

// TokenCache Token 缓存
type TokenCache struct {
	Token      string // Token
	AcquiredAt int64  // 获取时间（秒）
	ExpiresAt  int64  // 过期时间（秒）
}

// TokenResult Token 结果
type TokenResult struct {
	AppID      int64  // 应用ID
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
