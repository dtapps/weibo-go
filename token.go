package weibo

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// TokenCache Token 缓存
type TokenCache struct {
	token      string
	acquiredAt int64
	expiresIn  int
}

// TokenManager Token 管理器
type TokenManager struct {
	mu         sync.RWMutex
	cache      *TokenCache
	httpClient *HTTPClient
}

// NewTokenManager 创建 Token 管理器
func NewTokenManager() *TokenManager {
	return &TokenManager{
		httpClient: NewHTTPClient(),
	}
}

// isTokenValid 检查 token 是否有效
func (tm *TokenManager) isTokenValid() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.cache == nil {
		return false
	}

	expiresAt := tm.cache.acquiredAt + int64(tm.cache.expiresIn)*1000 - int64(TokenRefreshBufferSeconds)*1000
	return time.Now().UnixMilli() < expiresAt
}

// GetValidToken 获取有效的 token
func (tm *TokenManager) GetValidToken(appId, appSecret, tokenEndpoint string) (string, error) {
	// 检查缓存的 token 是否有效
	if tm.isTokenValid() {
		tm.mu.RLock()
		token := tm.cache.token
		tm.mu.RUnlock()
		return token, nil
	}

	// 获取新 token
	result, err := tm.FetchToken(appId, appSecret, tokenEndpoint)
	if err != nil {
		return "", err
	}

	return result.Token, nil
}

// FetchToken 获取 token
func (tm *TokenManager) FetchToken(appId, appSecret, tokenEndpoint string) (*TokenResult, error) {
	if tokenEndpoint == "" {
		tokenEndpoint = DefaultTokenEndpoint
	}

	var lastErr error

	for attempt := 0; attempt <= TokenFetchMaxRetries; attempt++ {
		var response TokenResponse
		err := tm.httpClient.PostJSON(tokenEndpoint, map[string]string{
			"app_id":     appId,
			"app_secret": appSecret,
		}, &response)

		if err != nil {
			lastErr = err

			// 检查是否是网络错误，需要重试
			if attempt < TokenFetchMaxRetries {
				delay := min(time.Duration(TokenFetchBaseDelay)*time.Duration(1<<attempt), TokenFetchMaxDelay)
				time.Sleep(delay)
				continue
			}

			return nil, fmt.Errorf("获取微博 token 失败: %w", err)
		}

		if response.Data.Token == "" {
			lastErr = fmt.Errorf("获取微博 token 失败: 响应中缺少 token")
			if attempt < TokenFetchMaxRetries {
				delay := min(time.Duration(TokenFetchBaseDelay)*time.Duration(1<<attempt), TokenFetchMaxDelay)
				time.Sleep(delay)
				continue
			}
			return nil, lastErr
		}

		result := &TokenResult{
			Token:      response.Data.Token,
			ExpiresIn:  response.Data.ExpireIn,
			AcquiredAt: time.Now().UnixMilli(),
			Uid:        response.Data.Uid,
		}

		// 缓存 token
		tm.mu.Lock()
		tm.cache = &TokenCache{
			token:      result.Token,
			acquiredAt: result.AcquiredAt,
			expiresIn:  result.ExpiresIn,
		}
		tm.mu.Unlock()

		return result, nil
	}

	return nil, lastErr
}

// GetCachedToken 获取缓存的 token 信息
func (tm *TokenManager) GetCachedToken() *TokenCache {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.cache == nil {
		return nil
	}

	return &TokenCache{
		token:      tm.cache.token,
		acquiredAt: tm.cache.acquiredAt,
		expiresIn:  tm.cache.expiresIn,
	}
}

// ClearCache 清除 token 缓存
func (tm *TokenManager) ClearCache() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.cache = nil
}

// GenerateMessageId 生成消息 ID
func GenerateMessageId() string {
	timestamp := time.Now().UnixMilli()
	randomStr := fmt.Sprintf("%d_%d", timestamp, time.Now().UnixNano()%1000000)
	hash := sha1.Sum([]byte(randomStr))
	return fmt.Sprintf("msg_%s", hex.EncodeToString(hash[:8]))
}

// ResolveInboundMessageId 解析入站消息 ID
func ResolveInboundMessageId(event *InboundMessage) string {
	payload := event.Payload

	// 如果有明确的 messageId
	if payload.MessageId != "" {
		return payload.MessageId
	}

	// 生成基于内容的哈希 ID
	var sb strings.Builder
	sb.WriteString(payload.FromUserId)
	if payload.Text != nil {
		sb.WriteString(*payload.Text)
	}
	if payload.Timestamp != nil {
		sb.WriteString(fmt.Sprintf("%d", *payload.Timestamp))
	}

	data, _ := json.Marshal(payload.Input)
	hasher := sha1.New()
	hasher.Write([]byte(sb.String()))
	hasher.Write(data)
	hash := hex.EncodeToString(hasher.Sum(nil))

	return fmt.Sprintf("weibo_inbound_%s", hash[:16])
}

// NormalizeChunkId 规范化分片 ID
func NormalizeChunkId(chunkId any) int {
	if chunkId == nil {
		return 0
	}

	switch v := chunkId.(type) {
	case float64:
		if v < 0 || v != v { // v != v 是 NaN 检查
			return 0
		}
		return int(v)
	case int:
		if v < 0 {
			return 0
		}
		return v
	case int64:
		if v < 0 {
			return 0
		}
		return int(v)
	default:
		return 0
	}
}

// NormalizeWeiboTarget 规范化微博目标
func NormalizeWeiboTarget(target string) string {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "user:") {
		return trimmed
	}
	return "user:" + trimmed
}

// LooksLikeWeiboId 检查是否像微博 ID
func LooksLikeWeiboId(str string) bool {
	str = strings.TrimSpace(str)
	if str == "" {
		return false
	}
	for _, c := range str {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// FormatWeiboTarget 格式化微博目标
func FormatWeiboTarget(userId string) string {
	return fmt.Sprintf("user:%s", userId)
}
