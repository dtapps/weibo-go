package token

import (
	"sync"
	"testing"
	"time"

	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/types"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.SetLevel(logger.LevelError)
}

// TestManager_ThreadSafety 验证 Token 管理器在高并发获取/刷新 Token 时的安全性
func TestManager_ThreadSafety(t *testing.T) {

	tm := NewManager()
	wg := sync.WaitGroup{}
	workerCount := 50

	wg.Add(workerCount)
	for range workerCount {
		go func() {
			defer wg.Done()
			// 模拟多个协程同时尝试获取缓存
			tm.GetCachedToken()
			// 模拟并发清除和检查
			tm.ClearCache()
			tm.isTokenValid()
		}()
	}
	wg.Wait()
}

// TestGetManager_ThreadSafety 验证全局 Map 容器的并发安全性
// 运行 go test -race 检查是否依然存在 member 包中出现的 map 竞争问题
func TestGetManager_ThreadSafety(t *testing.T) {

	wg := sync.WaitGroup{}
	accounts := []string{"acc_1", "acc_2", "acc_1", "acc_3"} // 包含重复 ID

	wg.Add(len(accounts) * 20)
	for _, accID := range accounts {
		for range 20 {
			go func(id string) {
				defer wg.Done()
				mgr := GetManager(id)
				assert.NotNil(t, mgr)
			}(accID)
		}
	}
	wg.Wait()
}

// TestManager_TokenExpirationLogic 验证 Token 有效期判定逻辑
func TestManager_TokenExpirationLogic(t *testing.T) {

	tm := NewManager()

	// 测试空缓存
	assert.False(t, tm.isTokenValid(), "空缓存应视为无效")

	// 测试未过期 Token
	tm.mu.Lock()
	tm.cache = &types.TokenCache{
		Token:      "fresh_token",
		AcquiredAt: time.Now().Unix(),
		ExpiresAt:  time.Now().Unix() + 3600, // 1小时后过期
	}
	tm.mu.Unlock()
	assert.True(t, tm.isTokenValid(), "有效期内的 Token 应视为有效")

	// 测试已过期 Token
	tm.mu.Lock()
	tm.cache.ExpiresAt = time.Now().Unix() - 10 // 已过期 10 秒
	tm.mu.Unlock()
	assert.False(t, tm.isTokenValid(), "过期的 Token 应视为无效")

	// 测试触发刷新缓冲区的 Token
	// 假设缓冲区是 60s，如果还有 30s 过期，isTokenValid 应返回 false
	tm.mu.Lock()
	tm.cache.ExpiresAt = time.Now().Unix() + int64((types.TokenRefreshBuffer/2)/time.Second)
	tm.mu.Unlock()
	assert.False(t, tm.isTokenValid(), "进入刷新缓冲区的 Token 应视为无效")
}

// TestManager_CacheOperations 验证缓存的基本操作
func TestManager_CacheOperations(t *testing.T) {

	tm := NewManager()

	// 设置缓存
	tm.mu.Lock()
	tm.cache = &types.TokenCache{Token: "test_data", ExpiresAt: 9999999999}
	tm.mu.Unlock()

	// 获取并验证
	cached := tm.GetCachedToken()
	assert.Equal(t, "test_data", cached.Token)

	// 清除并验证
	tm.ClearCache()
	assert.Nil(t, tm.GetCachedToken(), "清除后缓存应为 nil")
}
