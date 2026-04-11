package weibo

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"sync"
)

// ClientCacheEntry 客户端缓存条目
type ClientCacheEntry struct {
	Client      *WebSocketClient
	Fingerprint string
}

// ClientManager 客户端管理器
type ClientManager struct {
	mu           sync.RWMutex
	clients      map[string]*ClientCacheEntry
	tokenManager *TokenManager
}

// NewClientManager 创建客户端管理器
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:      make(map[string]*ClientCacheEntry),
		tokenManager: NewTokenManager(),
	}
}

// getFingerprint 获取连接指纹
func (m *ClientManager) getFingerprint(account *ResolvedAccount) string {
	var sb strings.Builder
	sb.WriteString(account.AccountId)
	if account.AppId != nil {
		sb.WriteString(*account.AppId)
	}
	if account.AppSecret != nil {
		sb.WriteString(*account.AppSecret)
	}
	sb.WriteString(account.WSEndpoint)
	sb.WriteString(account.TokenEndpoint)

	hasher := sha1.New()
	hasher.Write([]byte(sb.String()))
	return hex.EncodeToString(hasher.Sum(nil))
}

// CreateClient 创建或获取客户端
func (m *ClientManager) CreateClient(account *ResolvedAccount, options *WebSocketClientOptions) *WebSocketClient {
	fingerprint := m.getFingerprint(account)

	m.mu.RLock()
	cached, exists := m.clients[account.AccountId]
	m.mu.RUnlock()

	if exists && cached.Fingerprint == fingerprint {
		// 应用新选项
		if options != nil {
			if options.OnMessage != nil {
				cached.Client.OnMessage(options.OnMessage)
			}
			if options.OnError != nil {
				cached.Client.OnError(options.OnError)
			}
			if options.OnClose != nil {
				cached.Client.OnClose(options.OnClose)
			}
			if options.OnOpen != nil {
				cached.Client.OnOpen(options.OnOpen)
			}
			if options.OnStatus != nil {
				cached.Client.OnStatus(options.OnStatus)
			}
		}
		return cached.Client
	}

	// 关闭旧客户端
	if exists {
		cached.Client.Close()
		m.tokenManager.ClearCache()
	}

	// 创建新客户端
	client := NewWebSocketClient(account, options, m.tokenManager)

	m.mu.Lock()
	m.clients[account.AccountId] = &ClientCacheEntry{
		Client:      client,
		Fingerprint: fingerprint,
	}
	m.mu.Unlock()

	return client
}

// ClearClient 清除客户端
func (m *ClientManager) ClearClient(accountId string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cached, exists := m.clients[accountId]; exists {
		cached.Client.Close()
		delete(m.clients, accountId)
	}
}

// ClearAllClients 清除所有客户端
func (m *ClientManager) ClearAllClients() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cached := range m.clients {
		cached.Client.Close()
	}
	m.clients = make(map[string]*ClientCacheEntry)
}

// GetTokenManager 获取 Token 管理器
func (m *ClientManager) GetTokenManager() *TokenManager {
	return m.tokenManager
}

// ============ Account Resolution ============

const defaultAccountId = "default"

// ResolveAccount 解析账号配置
func ResolveAccount(config *WeiboConfig, accountId string) *ResolvedAccount {
	if accountId == "" {
		accountId = defaultAccountId
	}

	isDefault := accountId == defaultAccountId

	// 读取顶层配置
	var topLevelAppId *string
	var topLevelAppSecret *string
	var topLevelWsEndpoint *string
	var topLevelTokenEndpoint *string

	if config != nil {
		if config.AppId != nil && *config.AppId != "" {
			topLevelAppId = config.AppId
		}
		if config.AppSecret != nil && *config.AppSecret != "" {
			topLevelAppSecret = config.AppSecret
		}
		if config.WSEndpoint != nil && *config.WSEndpoint != "" {
			topLevelWsEndpoint = config.WSEndpoint
		}
		if config.TokenEndpoint != nil && *config.TokenEndpoint != "" {
			topLevelTokenEndpoint = config.TokenEndpoint
		}
	}

	if isDefault && config != nil {
		hasCredentials := topLevelAppId != nil && topLevelAppSecret != nil
		enabled := true
		if config.Enabled != nil {
			enabled = *config.Enabled
		}

		wsEndpoint := DefaultWSEndpoint
		if topLevelWsEndpoint != nil {
			wsEndpoint = *topLevelWsEndpoint
		}

		tokenEndpoint := DefaultTokenEndpoint
		if topLevelTokenEndpoint != nil {
			tokenEndpoint = *topLevelTokenEndpoint
		}

		dmPolicy := "open"
		if config.DMPolicy != nil {
			dmPolicy = *config.DMPolicy
		}

		allowFrom := []string{}
		if config.AllowFrom != nil {
			for _, v := range config.AllowFrom {
				if s := ReadOptionalString(v); s != "" {
					allowFrom = append(allowFrom, s)
				}
			}
		}

		chunkMode := "raw"
		if config.ChunkMode != nil {
			chunkMode = *config.ChunkMode
		}

		blockStreaming := true
		if config.BlockStreaming != nil {
			blockStreaming = *config.BlockStreaming
		}

		name := "Default"

		return &ResolvedAccount{
			AccountId:     defaultAccountId,
			Enabled:       enabled,
			Configured:    hasCredentials,
			Name:          &name,
			AppId:         topLevelAppId,
			AppSecret:     topLevelAppSecret,
			WSEndpoint:    wsEndpoint,
			TokenEndpoint: tokenEndpoint,
			Config: AccountConfig{
				DMPolicy:       dmPolicy,
				AllowFrom:      allowFrom,
				TokenEndpoint:  tokenEndpoint,
				WSEndpoint:     wsEndpoint,
				TextChunkLimit: config.TextChunkLimit,
				ChunkMode:      chunkMode,
				BlockStreaming: &blockStreaming,
			},
		}
	}

	// 读取子账号配置
	var accountCfg map[string]any
	if config != nil && config.Accounts != nil {
		if acc, ok := config.Accounts[accountId].(map[string]any); ok {
			accountCfg = acc
		}
	}

	// 合并配置
	mergedAppId := readStringFromMap(accountCfg, "appId")
	if mergedAppId == "" && topLevelAppId != nil {
		mergedAppId = *topLevelAppId
	}

	mergedAppSecret := readStringFromMap(accountCfg, "appSecret")
	if mergedAppSecret == "" && topLevelAppSecret != nil {
		mergedAppSecret = *topLevelAppSecret
	}

	mergedWsEndpoint := readStringFromMap(accountCfg, "wsEndpoint")
	if mergedWsEndpoint == "" {
		if topLevelWsEndpoint != nil {
			mergedWsEndpoint = *topLevelWsEndpoint
		} else {
			mergedWsEndpoint = DefaultWSEndpoint
		}
	}

	mergedTokenEndpoint := readStringFromMap(accountCfg, "tokenEndpoint")
	if mergedTokenEndpoint == "" {
		if topLevelTokenEndpoint != nil {
			mergedTokenEndpoint = *topLevelTokenEndpoint
		} else {
			mergedTokenEndpoint = DefaultTokenEndpoint
		}
	}

	dmPolicy := readStringFromMap(accountCfg, "dmPolicy")
	if dmPolicy == "" {
		if config != nil && config.DMPolicy != nil {
			dmPolicy = *config.DMPolicy
		} else {
			dmPolicy = "open"
		}
	}

	allowFrom := []string{}
	if accountCfg != nil {
		if arr, ok := accountCfg["allowFrom"].([]any); ok {
			for _, v := range arr {
				if s := ReadOptionalString(v); s != "" {
					allowFrom = append(allowFrom, s)
				}
			}
		}
	}
	if len(allowFrom) == 0 && config != nil && config.AllowFrom != nil {
		for _, v := range config.AllowFrom {
			if s := ReadOptionalString(v); s != "" {
				allowFrom = append(allowFrom, s)
			}
		}
	}

	textChunkLimit := readIntFromMap(accountCfg, "textChunkLimit")
	if textChunkLimit == 0 && config != nil && config.TextChunkLimit != nil {
		textChunkLimit = *config.TextChunkLimit
	}

	chunkMode := readStringFromMap(accountCfg, "chunkMode")
	if chunkMode == "" {
		if config != nil && config.ChunkMode != nil {
			chunkMode = *config.ChunkMode
		} else {
			chunkMode = "raw"
		}
	}

	blockStreaming := true
	if config != nil && config.BlockStreaming != nil {
		blockStreaming = *config.BlockStreaming
	}
	if bc := readBoolFromMap(accountCfg, "blockStreaming"); bc != nil {
		blockStreaming = *bc
	}

	enabled := true
	if config != nil && config.Enabled != nil {
		enabled = *config.Enabled
	}
	if ec := readBoolFromMap(accountCfg, "enabled"); ec != nil {
		enabled = *ec
	}

	hasCredentials := mergedAppId != "" && mergedAppSecret != ""

	var name *string
	if n := readStringFromMap(accountCfg, "name"); n != "" {
		name = &n
	}

	var appIdPtr, appSecretPtr *string
	if mergedAppId != "" {
		appIdPtr = &mergedAppId
	}
	if mergedAppSecret != "" {
		appSecretPtr = &mergedAppSecret
	}

	return &ResolvedAccount{
		AccountId:     accountId,
		Enabled:       enabled,
		Configured:    hasCredentials,
		Name:          name,
		AppId:         appIdPtr,
		AppSecret:     appSecretPtr,
		WSEndpoint:    mergedWsEndpoint,
		TokenEndpoint: mergedTokenEndpoint,
		Config: AccountConfig{
			DMPolicy:       dmPolicy,
			AllowFrom:      allowFrom,
			TokenEndpoint:  mergedTokenEndpoint,
			WSEndpoint:     mergedWsEndpoint,
			TextChunkLimit: &textChunkLimit,
			ChunkMode:      chunkMode,
			BlockStreaming: &blockStreaming,
		},
	}
}

// ListAccountIds 列出所有账号 ID
func ListAccountIds(config *WeiboConfig) []string {
	ids := []string{defaultAccountId}
	if config != nil && config.Accounts != nil {
		for id := range config.Accounts {
			if id != defaultAccountId {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

// ListEnabledAccounts 列出所有启用的账号
func ListEnabledAccounts(config *WeiboConfig) []*ResolvedAccount {
	var accounts []*ResolvedAccount
	for _, id := range ListAccountIds(config) {
		account := ResolveAccount(config, id)
		if account.Enabled && account.Configured {
			accounts = append(accounts, account)
		}
	}
	return accounts
}

// 辅助函数
func readStringFromMap(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok {
		return ReadOptionalString(v)
	}
	return ""
}

func readIntFromMap(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		case int64:
			return int(val)
		}
	}
	return 0
}

func readBoolFromMap(m map[string]any, key string) *bool {
	if m == nil {
		return nil
	}
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case bool:
			return &val
		}
	}
	return nil
}
