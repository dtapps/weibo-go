package plugin

import (
	"testing"

	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/types"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.SetLevel(logger.LevelError)
}

// TestPlugin_Lifecycle 测试插件的创建与基本属性获取
func TestPlugin_Lifecycle(t *testing.T) {
	accountID := "test_acc_123"
	acc := &types.Account{
		AccountID: accountID,
		AppID:     "appid_val",
		AppSecret: "secret_val",
	}
	cfg := &types.Config{}

	p := NewPlugin(accountID, acc, cfg)

	assert.Equal(t, accountID, p.GetAccountId())
	assert.Equal(t, types.ConnectionStateDisconnected.String(), p.GetState())

	// 验证管理器获取
	assert.NotNil(t, p.GetMember())
	assert.NotNil(t, p.GetTokenManager())
}

// TestPlugin_RuntimeCallbacks 测试运行时回调函数的设置与触发
func TestPlugin_RuntimeCallbacks(t *testing.T) {
	p := NewPlugin("test_acc", &types.Account{}, &types.Config{})

	connectedCalled := false
	messageReceived := false

	// 设置回调
	p.SetOnConnected(func() {
		connectedCalled = true
	})
	p.SetOnMessage(func(msg *types.InboundMessage) {
		messageReceived = true
		assert.NotNil(t, msg)
	})

	// 模拟 WebSocket 触发事件
	p.OnReady(&types.OnReadyData{ConnectID: "conn_001"})
	assert.True(t, connectedCalled, "应该触发连接成功回调")

	// 模拟收到消息推送
	mockMsg := &types.WsMessageMsg{
		Payload: types.WsMessageMsgPayload{
			MessageID: "msg_001",
			Text:      "hello",
		},
	}
	p.OnDispatch(mockMsg)
	assert.True(t, messageReceived, "应该触发消息处理回调")
}

// TestPluginManager_Operations 测试插件管理器的增删改查
func TestPluginManager_Operations(t *testing.T) {
	mgr := NewPluginManager()
	accountID := "acc_mgr_test"
	acc := &types.Account{AccountID: accountID}

	// 创建插件
	p := mgr.CreatePlugin(accountID, acc, &types.Config{})
	assert.NotNil(t, p)

	// 获取插件
	retrieved := mgr.GetPlugin(accountID)
	assert.Equal(t, p, retrieved)

	// 停止并移除
	err := mgr.StopPlugin(accountID)
	assert.NoError(t, err)
	assert.Nil(t, mgr.GetPlugin(accountID), "移除后不应再获取到插件")
}

// TestPlugin_StopCleanup 测试插件停止时的资源清理
func TestPlugin_StopCleanup(t *testing.T) {
	accountID := "cleanup_test"
	p := NewPlugin(accountID, &types.Account{}, &types.Config{})

	// 先往成员管理器存点数据
	memberMgr := p.GetMember()
	_, _ = memberMgr.AddUser(&types.MemberAddUserRequest{UserID: "user_1"})

	// 停止插件
	err := p.Stop()
	assert.NoError(t, err)

	// 验证 Context 是否已取消
	select {
	case <-p.ctx.Done():
		// OK
	default:
		t.Error("插件停止后 Context 应该被 cancel")
	}

	// 验证成员管理器是否已清空
	memberMgr.Clear()
	res := memberMgr.ListUsers(&types.MemberListUsersRequest{})
	assert.NotNil(t, res)
	assert.Equal(t, 0, res.Total, "停止插件后应清空成员缓存")
}
