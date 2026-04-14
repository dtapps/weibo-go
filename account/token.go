package account

import (
	"fmt"
	"sync"
	"time"

	"github.com/dtapps/weibo-go/http"
	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/types"
)

// Manager 账号管理器
type Manager struct {
	mu       sync.RWMutex
	accounts map[string]*types.ResolvedAccount
	log      *logger.Logger
}

var (
	globalManager *Manager
	managerOnce   sync.Once
)

// GetManager 获取全局账号管理器
func GetManager() *Manager {
	managerOnce.Do(func() {
		globalManager = NewManager()
	})
	return globalManager
}

// NewManager 创建账号管理器
func NewManager() *Manager {
	return &Manager{
		accounts: make(map[string]*types.ResolvedAccount),
		log:      logger.New("account"),
	}
}

// TokenManager Token 管理器
type TokenManager struct {
	mu         sync.RWMutex
	cache      *types.TokenCache
	httpClient *http.Client
}

// NewTokenManager 创建 Token 管理器
func NewTokenManager() *TokenManager {
	return &TokenManager{
		httpClient: http.NewClient(),
	}
}

// isTokenValid 检查 Token 是否有效
func (tm *TokenManager) isTokenValid() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.cache == nil {
		return false
	}

	expiresAt := tm.cache.AcquiredAt + int64(tm.cache.ExpiresIn)*1000 - int64(types.TokenRefreshBufferSeconds)*1000
	return time.Now().UnixMilli() < expiresAt
}

// GetValidToken 获取有效的 Token（自动刷新）
func (tm *TokenManager) GetValidToken(appId string, appSecret string, tokenEndpoint string) (string, error) {
	if tm.isTokenValid() {
		tm.mu.RLock()
		token := tm.cache.Token
		tm.mu.RUnlock()
		return token, nil
	}

	result, err := tm.FetchToken(appId, appSecret, tokenEndpoint)
	if err != nil {
		return "", err
	}

	return result.Token, nil
}

// FetchToken 获取新的 Token
func (tm *TokenManager) FetchToken(appId string, appSecret string, tokenEndpoint string) (*types.TokenResult, error) {
	if tokenEndpoint == "" {
		tokenEndpoint = types.DefaultTokenEndpoint
	}

	var lastErr error

	for attempt := 0; attempt <= types.TokenFetchMaxRetries; attempt++ {
		var response types.TokenResponse
		err := tm.httpClient.PostJSON(tokenEndpoint, map[string]string{
			"app_id":     appId,
			"app_secret": appSecret,
		}, &response)

		if err != nil {
			lastErr = err

			if attempt < types.TokenFetchMaxRetries {
				delay := min(time.Duration(types.TokenFetchBaseDelay)*time.Duration(1<<attempt), types.TokenFetchMaxDelay)
				time.Sleep(delay)
				continue
			}

			return nil, fmt.Errorf("获取微博 token 失败: %w", err)
		}

		if response.Data.Token == "" {
			lastErr = fmt.Errorf("获取微博 token 失败: 响应中缺少 token")
			if attempt < types.TokenFetchMaxRetries {
				delay := min(time.Duration(types.TokenFetchBaseDelay)*time.Duration(1<<attempt), types.TokenFetchMaxDelay)
				time.Sleep(delay)
				continue
			}
			return nil, lastErr
		}

		result := &types.TokenResult{
			Token:      response.Data.Token,
			ExpiresIn:  response.Data.ExpireIn,
			AcquiredAt: time.Now().UnixMilli(),
			Uid:        response.Data.Uid,
		}

		tm.mu.Lock()
		tm.cache = &types.TokenCache{
			Token:      result.Token,
			AcquiredAt: result.AcquiredAt,
			ExpiresIn:  result.ExpiresIn,
		}
		tm.mu.Unlock()

		return result, nil
	}

	return nil, lastErr
}

// GetCachedToken 获取缓存的 Token
func (tm *TokenManager) GetCachedToken() *types.TokenCache {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.cache == nil {
		return nil
	}

	return &types.TokenCache{
		Token:      tm.cache.Token,
		AcquiredAt: tm.cache.AcquiredAt,
		ExpiresIn:  tm.cache.ExpiresIn,
	}
}

// ClearCache 清除缓存的 Token
func (tm *TokenManager) ClearCache() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.cache = nil
}

// ResolveAccount 解析账号配置
func (m *Manager) ResolveAccount(cfg *types.Config, accountId string) *types.ResolvedAccount {
	account := &types.ResolvedAccount{
		AccountID:  accountId,
		Enabled:    true,
		Configured: false,
	}

	if cfg != nil && cfg.Weibo != nil {
		weiboConfig := cfg.Weibo

		if weiboConfig.Enabled != nil {
			account.Enabled = *weiboConfig.Enabled
		}

		account.WsGatewayUrl = types.DefaultWSEndpoint
		if weiboConfig.WSEndpoint != "" {
			account.WsGatewayUrl = weiboConfig.WSEndpoint
		}

		account.TokenEndpoint = types.DefaultTokenEndpoint
		if weiboConfig.TokenEndpoint != "" {
			account.TokenEndpoint = weiboConfig.TokenEndpoint
		}

		account.AppId = weiboConfig.AppId
		account.AppSecret = weiboConfig.AppSecret

		if account.AppId != "" {
			account.Configured = true
		}

		account.Config.DMPolicy = "open"
		if weiboConfig.DMPolicy != "" {
			account.Config.DMPolicy = weiboConfig.DMPolicy
		}

		account.Config.ChunkMode = "raw"
		if weiboConfig.ChunkMode != "" {
			account.Config.ChunkMode = weiboConfig.ChunkMode
		}
	}

	return account
}
