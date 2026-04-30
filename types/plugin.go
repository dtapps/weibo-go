package types

const (
	Version = "1.0.7"
)

// OnReadyData 连接就绪
type OnReadyData struct {
	ConnectID string // 连接编号
	Timestamp int64  // 连接时间
}

// WsAuthData WebSocket认证信息
type WsAuthData struct {
	AppID   string // 应用ID
	Token   string // 令牌
	Version string // 版本号
}
