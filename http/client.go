package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/dtapps/weibo-go/logger"
)

type Client struct {
	client *http.Client
	log    *logger.Logger
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: logger.New("http"),
	}
}

func (c *Client) GetJSON(url string, result any) error {
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
		respBody, _ := io.ReadAll(resp.Body)
		c.log.Error("请求失败",
			logger.F("url", url),
			logger.F("status_code", resp.StatusCode),
			logger.F("status_text", resp.Status),
			logger.F("body", string(respBody)),
		)
		return fmt.Errorf("请求失败: %d %s - %s", resp.StatusCode, resp.Status, string(respBody))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) PostJSON(url string, body any, result any) error {
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
		c.log.Error("请求失败",
			logger.F("url", url),
			logger.F("status_code", resp.StatusCode),
			logger.F("status_text", resp.Status),
			logger.F("body", string(respBody)),
		)
		return fmt.Errorf("请求失败: %d %s - %s", resp.StatusCode, resp.Status, string(respBody))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

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
