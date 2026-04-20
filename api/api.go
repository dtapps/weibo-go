package api

import (
	"fmt"

	"github.com/dtapps/weibo-go/http"
	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/types"
)

// API
type API struct {
	client *http.Client
	log    *logger.Logger
	token  string
}

// NewAPI 创建API实例
func NewAPI() *API {
	return &API{
		client: http.NewClient(),
		log:    logger.New("api"),
	}
}

// SetToken 设置 Token
func (a *API) SetToken(token string) {
	a.token = token
}

// GetHotSearch 获取热搜榜
func (a *API) GetHotSearch(params *types.HotSearchRequest) (*types.HotSearchResponse, error) {
	// URL
	url := types.DefaultHotSearchEndpoint

	// 验证榜单类型
	_, ok := types.CategoryMap[params.Category]
	if !ok {
		return nil, fmt.Errorf("无效的榜单类型: %s。支持的类型: 主榜、文娱榜、社会榜、生活榜、acg榜、科技榜、体育榜", params.Category)
	}

	// 参数
	queryParams := make(map[string]any)
	if params != nil && params.Category != "" {
		queryParams["category"] = params.Category
	}
	if a.token != "" {
		queryParams["token"] = a.token
	}
	if len(queryParams) > 0 {
		url = http.AddQueryParams(url, queryParams)
	}

	var response types.HotSearchResponse
	err := a.client.GetJSON(url, &response)
	if err != nil {
		return nil, err
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("获取热搜失败: %s", response.Message)
	}

	return &response, nil
}

// Search 搜索微博
func (a *API) Search(params *types.SearchRequest) (*types.SearchResponse, error) {
	// URL
	url := types.DefaultSearchEndpoint

	// 参数
	queryParams := map[string]any{
		"query": params.Query,
	}
	if a.token != "" {
		queryParams["token"] = a.token
	}
	if len(queryParams) > 0 {
		url = http.AddQueryParams(url, queryParams)
	}

	var response types.SearchResponse
	err := a.client.GetJSON(url, &response)
	if err != nil {
		return nil, fmt.Errorf("搜索微博失败: %s", response.Message)
	}

	return &response, nil
}

// GetUserStatus 获取用户微博
func (a *API) GetUserStatus(params *types.UserStatusRequest) (*types.UserStatusResponse, error) {
	// URL
	url := types.DefaultStatusEndpoint

	// 参数
	queryParams := make(map[string]any)
	if params != nil && params.Count != nil {
		queryParams["count"] = *params.Count
	}
	if a.token != "" {
		queryParams["token"] = a.token
	}
	if len(queryParams) > 0 {
		url = http.AddQueryParams(url, queryParams)
	}

	var response types.UserStatusResponse
	err := a.client.GetJSON(url, &response)
	if err != nil {
		return nil, fmt.Errorf("获取用户微博失败: %s", response.Message)
	}

	return &response, nil
}
