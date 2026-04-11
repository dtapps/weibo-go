package weibo

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ============ Hot Search API ============

// HotSearchResult 热搜结果
type HotSearchResult struct {
	Success  bool            `json:"success"`
	Category string          `json:"category,omitempty"`
	Total    int             `json:"total,omitempty"`
	CallTime string          `json:"callTime,omitempty"`
	Source   string          `json:"source,omitempty"`
	Items    []HotSearchItem `json:"items,omitempty"`
	Message  string          `json:"message,omitempty"`
	Error    string          `json:"error,omitempty"`
}

// FetchHotSearch 获取热搜榜
func FetchHotSearch(token string, category string, count *int, endpoint string) (*HotSearchResult, error) {
	if endpoint == "" {
		endpoint = DefaultHotSearchEndpoint
	}

	// 验证榜单类型
	_, ok := CategoryMap[category]
	if !ok {
		return &HotSearchResult{
			Success: false,
			Error:   fmt.Sprintf("无效的榜单类型: %s。支持的类型: 主榜、文娱榜、社会榜、生活榜、acg榜、科技榜、体育榜", category),
		}, nil
	}

	// 构建 URL
	params := map[string]any{
		"token":    token,
		"category": category,
	}
	if count != nil && *count > 0 {
		params["count"] = *count
	}

	url := AddQueryParams(endpoint, params)

	var response HotSearchResponse
	if err := httpGetJSON(url, &response); err != nil {
		return nil, fmt.Errorf("获取热搜榜失败: %w", err)
	}

	if response.Code != 0 {
		return &HotSearchResult{
			Success: false,
			Error:   response.Message,
		}, nil
	}

	data := response.Data

	if len(data.Data) == 0 {
		callTime := ""
		if data.CallTime != nil {
			callTime = *data.CallTime
		}
		source := ""
		if data.Source != nil {
			source = *data.Source
		}
		return &HotSearchResult{
			Success:  true,
			Category: category,
			Total:    0,
			CallTime: callTime,
			Source:   source,
			Items:    []HotSearchItem{},
			Message:  "没有找到热搜内容",
		}, nil
	}

	items := make([]HotSearchItem, len(data.Data))
	for i, item := range data.Data {
		items[i] = HotSearchItem{
			Rank:     item.Id,
			Word:     item.Word,
			HotValue: item.Num,
			Category: item.Cat,
			Flag:     item.Flag,
			AppLink:  item.AppQueryLink,
			H5Link:   item.H5QueryLink,
			FlagIcon: item.FlagLink,
		}
	}

	callTime := ""
	if data.CallTime != nil {
		callTime = *data.CallTime
	}
	source := ""
	if data.Source != nil {
		source = *data.Source
	}

	return &HotSearchResult{
		Success:  true,
		Category: category,
		Total:    len(items),
		CallTime: callTime,
		Source:   source,
		Items:    items,
	}, nil
}

// ============ Search API ============

// SearchResult 搜索结果
type SearchResult struct {
	Success        bool   `json:"success"`
	Completed      bool   `json:"completed,omitempty"`
	Analyzing      bool   `json:"analyzing,omitempty"`
	Content        string `json:"content,omitempty"`
	ContentFormat  string `json:"contentFormat,omitempty"`
	ReferenceCount int    `json:"referenceCount,omitempty"`
	Scheme         string `json:"scheme,omitempty"`
	Version        string `json:"version,omitempty"`
	CallTime       string `json:"callTime,omitempty"`
	Source         string `json:"source,omitempty"`
	Message        string `json:"message,omitempty"`
	Error          string `json:"error,omitempty"`
	NoContent      bool   `json:"noContent,omitempty"`
}

// SearchWeibo 搜索微博
func SearchWeibo(query, token, endpoint string) (*SearchResult, error) {
	if endpoint == "" {
		endpoint = DefaultSearchEndpoint
	}

	params := map[string]any{
		"query": query,
		"token": token,
	}

	url := AddQueryParams(endpoint, params)

	var response SearchResponse
	if err := httpGetJSON(url, &response); err != nil {
		return nil, fmt.Errorf("微博搜索失败: %w", err)
	}

	if response.Code != 0 {
		return &SearchResult{
			Success: false,
			Error:   response.Message,
		}, nil
	}

	data := response.Data

	if data.NoContent {
		callTime := ""
		if data.CallTime != nil {
			callTime = *data.CallTime
		}
		source := ""
		if data.Source != nil {
			source = *data.Source
		}
		return &SearchResult{
			Success:   true,
			Completed: data.Completed,
			NoContent: true,
			CallTime:  callTime,
			Source:    source,
			Message:   "没有找到相关内容",
		}, nil
	}

	if data.Refused {
		return &SearchResult{
			Success: false,
			Error:   "搜索请求被拒绝",
		}, nil
	}

	callTime := ""
	if data.CallTime != nil {
		callTime = *data.CallTime
	}
	source := ""
	if data.Source != nil {
		source = *data.Source
	}

	return &SearchResult{
		Success:        true,
		Completed:      data.Completed,
		Analyzing:      data.Analyzing,
		Content:        data.Msg,
		ContentFormat:  data.MsgFormat,
		ReferenceCount: data.ReferenceNum,
		Scheme:         data.Scheme,
		Version:        data.Version,
		CallTime:       callTime,
		Source:         source,
	}, nil
}

// ============ Status API ============

// StatusResult 状态结果
type StatusResult struct {
	Success  bool         `json:"success"`
	Total    int          `json:"total,omitempty"`
	Statuses []StatusItem `json:"statuses,omitempty"`
	Message  string       `json:"message,omitempty"`
	Error    string       `json:"error,omitempty"`
}

// FetchWeiboStatus 获取用户微博
func FetchWeiboStatus(token string, count *int, endpoint string) (*StatusResult, error) {
	if endpoint == "" {
		endpoint = DefaultStatusEndpoint
	}

	params := map[string]any{
		"token": token,
	}
	if count != nil && *count > 0 {
		params["count"] = *count
	}

	url := AddQueryParams(endpoint, params)

	var response StatusResponse
	if err := httpGetJSON(url, &response); err != nil {
		return nil, fmt.Errorf("获取用户微博失败: %w", err)
	}

	if response.Code != 0 {
		return &StatusResult{
			Success: false,
			Error:   response.Message,
		}, nil
	}

	data := response.Data

	if len(data.Statuses) == 0 {
		return &StatusResult{
			Success:  true,
			Total:    0,
			Statuses: []StatusItem{},
			Message:  "没有找到微博内容",
		}, nil
	}

	return &StatusResult{
		Success:  true,
		Total:    data.TotalNumber,
		Statuses: data.Statuses,
	}, nil
}

// ============ Helper Functions ============

func httpGetJSON(url string, result any) error {
	client := NewHTTPClient()
	return client.GetJSON(url, result)
}

// ToJSON 将结果转换为 JSON 字符串
func ToJSON(data any) string {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "JSON marshal error: %v"}`, err)
	}
	return string(jsonData)
}

// ReadOptionalString 读取可选字符串
func ReadOptionalString(value any) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		var trimmed strings.Builder
		for _, c := range v {
			if c > ' ' {
				trimmed.WriteString(string(c))
			}
		}
		return trimmed.String()
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	default:
		return ""
	}
}
