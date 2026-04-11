package weibo

// ConnectionState 连接状态
type ConnectionState string

const (
	ConnectionStateIdle       ConnectionState = "idle"
	ConnectionStateConnecting ConnectionState = "connecting"
	ConnectionStateConnected  ConnectionState = "connected"
	ConnectionStateBackoff    ConnectionState = "backoff"
	ConnectionStateError      ConnectionState = "error"
	ConnectionStateStopped    ConnectionState = "stopped"
)

// DisconnectInfo 断开连接信息
type DisconnectInfo struct {
	Code   int    `json:"code"`
	Reason string `json:"reason"`
	At     int64  `json:"at"`
}

// RuntimeStatusPatch 运行时状态更新
type RuntimeStatusPatch struct {
	Running           *bool            `json:"running,omitempty"`
	Connected         *bool            `json:"connected,omitempty"`
	ConnectionState   *ConnectionState `json:"connectionState,omitempty"`
	ReconnectAttempts *int             `json:"reconnectAttempts,omitempty"`
	NextRetryAt       *int64           `json:"nextRetryAt,omitempty"`
	LastConnectedAt   *int64           `json:"lastConnectedAt,omitempty"`
	LastDisconnect    *DisconnectInfo  `json:"lastDisconnect,omitempty"`
	LastError         *string          `json:"lastError,omitempty"`
	LastStartAt       *int64           `json:"lastStartAt,omitempty"`
	LastStopAt        *int64           `json:"lastStopAt,omitempty"`
	LastInboundAt     *int64           `json:"lastInboundAt,omitempty"`
	LastOutboundAt    *int64           `json:"lastOutboundAt,omitempty"`
}

// MessageContext 消息上下文
type MessageContext struct {
	MessageId  string `json:"messageId"`
	SenderId   string `json:"senderId"`
	Text       string `json:"text"`
	CreateTime *int64 `json:"createTime,omitempty"`
}

// SendResult 发送结果
type SendResult struct {
	MessageId string `json:"messageId"`
	ChatId    string `json:"chatId"`
	ChunkId   int    `json:"chunkId"`
	Done      bool   `json:"done"`
}

// AccountConfig 账号配置
type AccountConfig struct {
	DMPolicy       string   `json:"dmPolicy"`
	AllowFrom      []string `json:"allowFrom"`
	TokenEndpoint  string   `json:"tokenEndpoint"`
	WSEndpoint     string   `json:"wsEndpoint"`
	TextChunkLimit *int     `json:"textChunkLimit,omitempty"`
	ChunkMode      string   `json:"chunkMode"`
	BlockStreaming *bool    `json:"blockStreaming,omitempty"`
}

// ResolvedAccount 已解析的账号
type ResolvedAccount struct {
	AccountId     string        `json:"accountId"`
	Enabled       bool          `json:"enabled"`
	Configured    bool          `json:"configured"`
	Name          *string       `json:"name,omitempty"`
	AppId         *string       `json:"appId,omitempty"`
	AppSecret     *string       `json:"appSecret,omitempty"`
	WSEndpoint    string        `json:"wsEndpoint"`
	TokenEndpoint string        `json:"tokenEndpoint"`
	Config        AccountConfig `json:"config"`
}

// WeiboConfig 微博配置
type WeiboConfig struct {
	Enabled        *bool          `json:"enabled,omitempty"`
	AppId          *string        `json:"appId,omitempty"`
	AppSecret      *string        `json:"appSecret,omitempty"`
	WSEndpoint     *string        `json:"wsEndpoint,omitempty"`
	TokenEndpoint  *string        `json:"tokenEndpoint,omitempty"`
	DMPolicy       *string        `json:"dmPolicy,omitempty"`
	AllowFrom      []any          `json:"allowFrom,omitempty"`
	TextChunkLimit *int           `json:"textChunkLimit,omitempty"`
	ChunkMode      *string        `json:"chunkMode,omitempty"`
	BlockStreaming *bool          `json:"blockStreaming,omitempty"`
	Accounts       map[string]any `json:"accounts,omitempty"`
}

// TokenResponse Token 响应
type TokenResponse struct {
	Data TokenData `json:"data"`
}

// TokenData Token 数据
type TokenData struct {
	Token    string `json:"token"`
	ExpireIn int    `json:"expire_in"`
	Uid      int64  `json:"uid"`
}

// TokenResult Token 结果
type TokenResult struct {
	Token      string `json:"token"`
	ExpiresIn  int    `json:"expiresIn"`
	AcquiredAt int64  `json:"acquiredAt"`
	Uid        int64  `json:"uid"`
}

// HotSearchItem 热搜项
type HotSearchItem struct {
	Rank     int    `json:"rank"`
	Word     string `json:"word"`
	HotValue int    `json:"hotValue"`
	Category string `json:"category"`
	Flag     int    `json:"flag"`
	AppLink  string `json:"appLink"`
	H5Link   string `json:"h5Link"`
	FlagIcon string `json:"flagIcon"`
}

// HotSearchResponse 热搜响应
type HotSearchResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    HotSearchData `json:"data"`
}

// HotSearchData 热搜数据
type HotSearchData struct {
	CallTime *string            `json:"callTime,omitempty"`
	Source   *string            `json:"source,omitempty"`
	Data     []HotSearchItemRaw `json:"data"`
}

// HotSearchItemRaw 热搜项原始数据
type HotSearchItemRaw struct {
	Cat          string `json:"cat"`
	Id           int    `json:"id"`
	Word         string `json:"word"`
	Num          int    `json:"num"`
	Flag         int    `json:"flag"`
	AppQueryLink string `json:"app_query_link"`
	H5QueryLink  string `json:"h5_query_link"`
	FlagLink     string `json:"flag_link"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    SearchData `json:"data"`
}

// SearchData 搜索数据
type SearchData struct {
	Analyzing       bool    `json:"analyzing"`
	Completed       bool    `json:"completed"`
	Msg             string  `json:"msg"`
	MsgFormat       string  `json:"msg_format"`
	MsgJson         string  `json:"msg_json"`
	NoContent       bool    `json:"noContent"`
	ProfileImageURL string  `json:"profile_image_url"`
	ReferenceNum    int     `json:"reference_num"`
	Refused         bool    `json:"refused"`
	Scheme          string  `json:"scheme"`
	Status          int     `json:"status"`
	StatusStage     int     `json:"status_stage"`
	Version         string  `json:"version"`
	CallTime        *string `json:"callTime,omitempty"`
	Source          *string `json:"source,omitempty"`
}

// StatusItem 微博状态项
type StatusItem struct {
	Id             int64       `json:"id"`
	Mid            string      `json:"mid"`
	Text           string      `json:"text"`
	CreatedAt      string      `json:"created_at"`
	HasImage       bool        `json:"has_image"`
	Images         []string    `json:"images,omitempty"`
	PicNum         *int        `json:"pic_num,omitempty"`
	CommentsCount  int         `json:"comments_count"`
	RepostsCount   int         `json:"reposts_count"`
	AttitudesCount int         `json:"attitudes_count"`
	User           StatusUser  `json:"user"`
	Repost         *StatusItem `json:"repost,omitempty"`
}

// StatusUser 状态用户
type StatusUser struct {
	ScreenName string `json:"screen_name"`
}

// StatusResponse 用户微博响应
type StatusResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    StatusData `json:"data"`
}

// StatusData 状态数据
type StatusData struct {
	Statuses    []StatusItem `json:"statuses"`
	TotalNumber int          `json:"total_number"`
}

// WebSocketMessage WebSocket 消息
type WebSocketMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}

// SendMessagePayload 发送消息载荷
type SendMessagePayload struct {
	ToUserId  string `json:"toUserId"`
	Text      string `json:"text"`
	MessageId string `json:"messageId"`
	ChunkId   int    `json:"chunkId"`
	Done      bool   `json:"done"`
}

// InboundMessage 接收到的消息
type InboundMessage struct {
	Type    string                `json:"type"`
	Payload InboundMessagePayload `json:"payload"`
}

// InboundMessagePayload 接收消息载荷
type InboundMessagePayload struct {
	MessageId  string                    `json:"messageId"`
	FromUserId string                    `json:"fromUserId"`
	Text       *string                   `json:"text,omitempty"`
	Timestamp  *int64                    `json:"timestamp,omitempty"`
	Input      []InboundMessageInputItem `json:"input,omitempty"`
}

// InboundMessageInputItem 消息输入项
type InboundMessageInputItem struct {
	Type    string               `json:"type"`
	Role    string               `json:"role"`
	Content []InboundContentPart `json:"content"`
}

// InboundContentPart 内容部分
type InboundContentPart struct {
	Type     string                `json:"type"`
	Text     *string               `json:"text,omitempty"`
	Source   *InboundContentSource `json:"source,omitempty"`
	Filename *string               `json:"filename,omitempty"`
}

// InboundContentSource 内容来源
type InboundContentSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// HotSearchParams 热搜参数
type HotSearchParams struct {
	Category string `json:"category"`
	Count    *int   `json:"count,omitempty"`
}

// SearchParams 搜索参数
type SearchParams struct {
	Query string `json:"query"`
}

// StatusParams 状态参数
type StatusParams struct {
	Count *int `json:"count,omitempty"`
}

// ResultResponse 结果响应
type ResultResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}
