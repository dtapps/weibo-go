package types

// HotSearchRequest 获取热搜榜 请求
type HotSearchRequest struct {
	Category string `json:"category"`
	Count    *int   `json:"count,omitempty"`
}

// HotSearchResponse 获取热搜榜 响应
type HotSearchResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    HotSearchData `json:"data"`
}

type HotSearchData struct {
	CallTime *string            `json:"callTime,omitempty"`
	Source   *string            `json:"source,omitempty"`
	Data     []HotSearchItemRaw `json:"data"`
}

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

// 搜索微博 请求
type SearchRequest struct {
	Query string `json:"query"`
}

// 搜索微博 响应
type SearchResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    SearchData `json:"data"`
}

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

// 获取用户微博 请求
type UserStatusRequest struct {
	Count *int `json:"count,omitempty"`
}

// 获取用户微博 响应
type UserStatusResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    UserStatusData `json:"data"`
}

type UserStatusData struct {
	Statuses    []UserStatusItem `json:"statuses"`
	TotalNumber int              `json:"total_number"`
}

type UserStatusItem struct {
	Id             int64           `json:"id"`
	Mid            string          `json:"mid"`
	Text           string          `json:"text"`
	CreatedAt      string          `json:"created_at"`
	HasImage       bool            `json:"has_image"`
	Images         []string        `json:"images,omitempty"`
	PicNum         *int            `json:"pic_num,omitempty"`
	CommentsCount  int             `json:"comments_count"`
	RepostsCount   int             `json:"reposts_count"`
	AttitudesCount int             `json:"attitudes_count"`
	User           StatusUser      `json:"user"`
	Repost         *UserStatusItem `json:"repost,omitempty"`
}

type StatusUser struct {
	ScreenName string `json:"screen_name"`
}
