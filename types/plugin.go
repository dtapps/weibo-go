package types

// OnReadyData 连接就绪
type OnReadyData struct {
	ConnectID string `json:"connectId"` // 连接编号
	Timestamp int64  `json:"timestamp"` // 连接时间
}

// WsAuth WebSocket认证
type WsAuth struct {
	BizID string `json:"bizId"`
	UID   string `json:"uid"`
	Token string `json:"token"`
}
