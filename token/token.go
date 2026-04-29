package token

import (
	"fmt"
	"sync"
	"time"

	"github.com/dtapps/weibo-go/http"
	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/types"
)

// Manager Token 管理器
type Manager struct {
	mu         sync.RWMutex
	cache      *types.TokenCache
	httpClient *http.Client
	log        *logger.Logger
	callback   types.TokenCallback // Token 回调
}

// 全局 Token 管理器
var (
	globalManagers map[string]*Manager
	globalMu       sync.RWMutex
	managerOnce    sync.Once
)

// GetManager 获取指定账号的 Token 管理器
func GetManager(accountID string) *Manager {
	managerOnce.Do(func() {
		globalManagers = make(map[string]*Manager)
	})

	globalMu.RLock()
	mgr, ok := globalManagers[accountID]
	globalMu.RUnlock()
	if ok {
		return mgr
	}

	globalMu.Lock()
	defer globalMu.Unlock()

	if mgr, ok = globalManagers[accountID]; ok {
		return mgr
	}

	mgr = NewManager()
	globalManagers[accountID] = mgr
	return mgr
}

// NewManager 创建 Token 管理器
func NewManager() *Manager {
	return &Manager{
		cache:      nil,
		httpClient: http.NewClient(),
		log:        logger.New("token"),
		callback:   nil,
	}
}

// SetCallback 设置 Token 回调
func (m *Manager) SetCallback(callback types.TokenCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callback = callback
}

// GetCallback 获取 Token 回调
func (m *Manager) GetCallback() types.TokenCallback {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callback
}

// isTokenValid 检查 Token 是否有效
func (m *Manager) isTokenValid() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.cache == nil {
		return false
	}

	expiresAt := m.cache.ExpiresAt
	if expiresAt == 0 {
		expiresAt = m.cache.AcquiredAt + int64(types.TokenExpireDuration/time.Second)
	}

	expiresAt = expiresAt - int64(types.TokenRefreshBuffer/time.Second)

	now := time.Now().Unix()
	isValid := now < expiresAt
	return isValid
}

// IsTokenExpiringSoon 检查 Token 是否即将过期（在指定时间内过期）
func (m *Manager) IsTokenExpiringSoon(within time.Duration) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.cache == nil {
		return true
	}

	expiresAt := m.cache.ExpiresAt
	if expiresAt == 0 {
		expiresAt = m.cache.AcquiredAt + int64(types.TokenExpireDuration/time.Second)
	}

	now := time.Now().Unix()
	threshold := now + int64(within.Seconds())
	isExpiring := threshold >= expiresAt

	// m.log.Debug("检查 Token 过期",
	// 	logger.F("now", time.Unix(now, 0).Format("2006-01-02 15:04:05")),
	// 	logger.F("expiresAt", time.Unix(expiresAt, 0).Format("2006-01-02 15:04:05")),
	// 	logger.F("threshold", time.Unix(threshold, 0).Format("2006-01-02 15:04:05")),
	// 	logger.F("within", within.String()),
	// 	logger.F("baseTime", time.Unix(expiresAt, 0).Add(-within).Format("2006-01-02 15:04:05")),
	// 	logger.F("isExpiring", isExpiring),
	// )

	return isExpiring
}

// GetValidToken 获取有效的 Token（自动刷新）
func (m *Manager) GetValidToken(appID string, appSecret string, tokenEndpoint string) (string, error) {

	if m.isTokenValid() {
		m.mu.RLock()
		token := m.cache.Token
		m.mu.RUnlock()
		return token, nil
	}

	result, err := m.FetchToken(appID, appSecret, tokenEndpoint)
	if err != nil {
		return "", err
	}

	return result.Token, nil
}

// FetchToken 获取新的 Token
func (m *Manager) FetchToken(appID string, appSecret string, tokenEndpoint string) (*types.TokenResult, error) {

	if tokenEndpoint == "" {
		tokenEndpoint = types.DefaultTokenEndpoint
	}

	var lastErr error

	for attempt := 0; attempt <= types.TokenFetchMaxRetries; attempt++ {
		var response types.TokenResponse
		err := m.httpClient.PostJSON(tokenEndpoint, types.TokenRequest{
			AppID:     appID,
			AppSecret: appSecret,
		}, &response)

		if err != nil {
			lastErr = err

			if attempt < types.TokenFetchMaxRetries {
				delay := min(time.Duration(types.TokenFetchBaseDelay)*time.Duration(1<<attempt), types.TokenFetchMaxDelay)
				time.Sleep(delay)
				continue
			}

			// 调用回调（失败）
			m.callCallback(&types.TokenCallbackData{
				Status: "error",
				AppID:  appID,
				Error:  lastErr,
			})

			return nil, lastErr
		}

		if response.Data.Token == "" {
			lastErr = fmt.Errorf("响应中缺少 token")

			if attempt < types.TokenFetchMaxRetries {
				delay := min(time.Duration(types.TokenFetchBaseDelay)*time.Duration(1<<attempt), types.TokenFetchMaxDelay)
				time.Sleep(delay)
				continue
			}

			// 调用回调（失败）
			m.callCallback(&types.TokenCallbackData{
				Status: "error",
				AppID:  appID,
				Error:  lastErr,
			})

			return nil, lastErr
		}

		result := &types.TokenResult{
			AppID:      response.Data.Uid,      // 应用 ID
			Token:      response.Data.Token,    // Token
			AcquiredAt: time.Now().Unix(),      // 获取时间（秒）
			ExpiresIn:  response.Data.ExpireIn, // 过期时长（秒）
		}

		m.mu.Lock()
		m.cache = &types.TokenCache{
			Token:      result.Token,                         // Token
			AcquiredAt: result.AcquiredAt,                    // 获取时间（秒）
			ExpiresAt:  result.AcquiredAt + result.ExpiresIn, // 过期时长（秒）
		}
		m.mu.Unlock()

		// 调用回调（成功）
		m.callCallback(&types.TokenCallbackData{
			Status:     "success",
			AppID:      appID,
			Token:      result.Token,
			ExpiresIn:  result.ExpiresIn,
			AcquiredAt: result.AcquiredAt,
			ExpiresAt:  result.AcquiredAt + result.ExpiresIn,
		})

		return result, nil
	}

	return nil, lastErr
}

// callCallback 调用 Token 回调
func (m *Manager) callCallback(data *types.TokenCallbackData) {
	m.mu.RLock()
	callback := m.callback
	m.mu.RUnlock()

	if callback != nil {
		go callback(data)
	}
}

// GetCachedToken 获取缓存的 Token
func (m *Manager) GetCachedToken() *types.TokenCache {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.cache == nil {
		return nil
	}

	return &types.TokenCache{
		Token:      m.cache.Token,      // Token
		AcquiredAt: m.cache.AcquiredAt, // 过期时长（秒）
		ExpiresAt:  m.cache.ExpiresAt,  // 过期时间（秒）
	}
}

// ClearCache 清除缓存的 Token
func (m *Manager) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache = nil
}
