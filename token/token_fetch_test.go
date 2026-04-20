package token

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/types"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.SetLevel(logger.LevelError)
}

// TestManager_FetchToken_Success 测试正常的 Token 获取流程
// 验证网络请求发送、JSON 解析、结果映射以及内部缓存的同步更新
func TestManager_FetchToken_Success(t *testing.T) {
	tm := NewManager()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		resp := types.TokenResponse{
			Data: types.TokenData{
				Uid:      123456789,
				Token:    "mock_access_token",
				ExpireIn: 3600,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	result, err := tm.FetchToken("app_123", "secret_456", server.URL)

	// 验证返回的 Result 对象
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "mock_access_token", result.Token)
	assert.Equal(t, int64(123456789), result.AppID)

	// 验证 Manager 内部缓存是否已自动更新
	cached := tm.GetCachedToken()
	assert.NotNil(t, cached)
	assert.Equal(t, "mock_access_token", cached.Token)
}

// TestManager_FetchToken_RetryFailure 测试网络异常时的指数退避重试机制
// 验证当接口返回非 200 状态码时，FetchToken 是否按照预设次数进行重试
func TestManager_FetchToken_RetryFailure(t *testing.T) {
	tm := NewManager()
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusBadGateway) // 模拟网关错误
	}))
	defer server.Close()

	_, err := tm.FetchToken("app_1", "secret_1", server.URL)

	assert.Error(t, err)
	// 动态计算期望调用次数 (1次初始请求 + N次重试)
	expectedCalls := 1 + types.TokenFetchMaxRetries
	assert.Equal(t, expectedCalls, callCount)
}

// TestManager_FetchToken_MissingToken 测试空数据容错处理
// 验证当接口响应 200 但 Data 字段中没有 Token 时的错误拦截逻辑
func TestManager_FetchToken_MissingToken(t *testing.T) {
	tm := NewManager()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 模拟服务器返回了 Uid 但没有返回 Token 的异常场景
		resp := types.TokenResponse{
			Data: types.TokenData{
				Uid: 12345,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	_, err := tm.FetchToken("app_1", "secret_1", server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "响应中缺少 token")
}
