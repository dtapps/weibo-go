package account

import (
	"fmt"
	"sync"

	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/types"
)

// Manager 账号管理器
type Manager struct {
	mu       sync.RWMutex
	accounts map[string]*types.Account // accountID -> Account
	log      *logger.Logger
}

// 全局账号管理器
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
		accounts: make(map[string]*types.Account),
		log:      logger.New("account"),
	}
}

// ResolveAccount 解析账号配置
func (m *Manager) ResolveAccount(cfg *types.Config, accountID string) *types.Account {
	m.mu.Lock()
	defer m.mu.Unlock()

	account := &types.Account{
		AccountID: accountID,
	}

	// 合并配置
	if cfg != nil && cfg.Weibo != nil {
		weiboConfig := cfg.Weibo

		if weiboConfig.Enabled != nil {
			account.Enabled = *weiboConfig.Enabled
		}

		account.WSEndpoint = types.DefaultWSEndpoint
		if weiboConfig.WSEndpoint != "" {
			account.WSEndpoint = weiboConfig.WSEndpoint
		}

		account.TokenEndpoint = types.DefaultTokenEndpoint
		if weiboConfig.TokenEndpoint != "" {
			account.TokenEndpoint = weiboConfig.TokenEndpoint
		}

		account.AppID = weiboConfig.AppID
		account.AppSecret = weiboConfig.AppSecret
		if account.AppID != "" && account.AppSecret != "" {
			account.Configured = true
		}

		account.DMPolicy = "open"
		if weiboConfig.DMPolicy != "" {
			account.DMPolicy = weiboConfig.DMPolicy
		}

		// 分片模式 newline=按段落分片, length=按字符数分片, raw=转发上游分片
		account.ChunkMode = "raw"
		if weiboConfig.ChunkMode != "" {
			account.ChunkMode = weiboConfig.ChunkMode
		}

		account.WsMaxReconnectAttempts = 100

		account.Config = weiboConfig
	}

	m.log.Debug("解析账号",
		logger.F("accountID", accountID),
		logger.F("configured", account.Configured),
	)

	return account
}

// AddAccount 添加账号
func (m *Manager) AddAccount(accountID string, account *types.Account) (*types.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.accounts[accountID] = account

	return account, nil
}

// UpdateAccount 更新账号
func (m *Manager) UpdateAccount(accountID string, account *types.Account) (*types.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.accounts[accountID] = account

	return account, nil
}

// DeleteAccount 删除账号
func (m *Manager) DeleteAccount(accountID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.accounts, accountID)

	return nil
}

// ListAccounts 列出所有账号
func (m *Manager) ListAccounts() *types.AccountListAccountsResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()

	accounts := make([]*types.Account, 0, len(m.accounts))
	for _, account := range m.accounts {
		accounts = append(accounts, account)
	}

	return &types.AccountListAccountsResponse{
		Total:    len(accounts),
		Accounts: accounts,
	}
}

// GetAccount 获取账号
func (m *Manager) GetAccount(accountID string) (*types.Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if account, ok := m.accounts[accountID]; ok {
		return account, nil
	}

	return &types.Account{AccountID: accountID}, fmt.Errorf("account not found: %s", accountID)
}
