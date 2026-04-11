package weibo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// HTTPClient HTTP 客户端
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient 创建 HTTP 客户端
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetJSON 发送 GET 请求并解析 JSON 响应
func (c *HTTPClient) GetJSON(url string, result any) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("请求失败: %d %s - %s", resp.StatusCode, resp.Status, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// PostJSON 发送 POST 请求并解析 JSON 响应
func (c *HTTPClient) PostJSON(url string, body any, result any) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("请求失败: %d %s - %s", resp.StatusCode, resp.Status, string(respBody))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// AddQueryParams 添加查询参数到 URL
func AddQueryParams(baseURL string, params map[string]any) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	q := u.Query()
	for k, v := range params {
		if v == nil || v == "" {
			continue
		}
		q.Set(k, fmt.Sprintf("%v", v))
	}

	u.RawQuery = q.Encode()
	return u.String()
}
