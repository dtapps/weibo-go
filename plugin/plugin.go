package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dtapps/weibo-go/account"
	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/member"
	"github.com/dtapps/weibo-go/types"
	"github.com/dtapps/weibo-go/ws"
)

// Plugin 微博插件
type Plugin struct {
	name         string
	version      string
	accountId    string
	account      *types.ResolvedAccount
	cfg          *types.Config
	client       *ws.WsClient
	runtime      *Runtime
	log          *logger.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	tokenManager *account.TokenManager
}

// Runtime 运行时
type Runtime struct {
	channel        *ChannelRuntime
	onMessage      func(msg *types.WsMessageMsg)
	onConnected    func()
	onDisconnected func()
}

// ChannelRuntime 通道运行时
type ChannelRuntime struct {
	// 发送文本
	SendText func(text string) error
}

// NewPlugin 创建插件
func New(accountId string, acc *types.ResolvedAccount, cfg *types.Config) *Plugin {
	ctx, cancel := context.WithCancel(context.Background())

	return &Plugin{
		name:      "weibo",
		version:   "1.0.0",
		accountId: accountId,
		account:   acc,
		cfg:       cfg,
		log:       logger.New("plugin"),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start 启动插件
func (p *Plugin) Start() error {
	p.log.Info("启动微博插件", logger.F("accountId", p.accountId))
	p.log.Info("配置信息",
		logger.F("appId", maskString(p.account.AppId)),
		logger.F("appSecret", maskString(p.account.AppSecret)),
		logger.F("wsEndpoint", p.account.WsGatewayUrl),
	)

	// 创建Token管理器
	p.tokenManager = account.NewTokenManager()

	// 获取Token
	token, err := p.tokenManager.GetValidToken(
		p.account.AppId,
		p.account.AppSecret,
		p.account.TokenEndpoint,
	)
	if err != nil {
		return fmt.Errorf("获取Token失败: %w", err)
	}

	// 创建WS客户端
	p.client = ws.NewWsClient(p.account.WsGatewayUrl, p.account.AccountID, p.account.BotId, p)

	// 设置认证信息
	p.client.SetAuth(&types.WsAuth{
		BizID: "weibo",
		UID:   p.account.BotId,
		Token: token,
	})

	// 设置重连配置
	p.client.SetReconnectConfig(0, "1s,2s,5s,10s,30s,60s")

	// 启动客户端
	return p.client.Connect()
}

// Stop 停止插件
func (p *Plugin) Stop() error {
	p.log.Info("停止微博插件", logger.F("accountId", p.accountId))

	p.cancel()

	// 移除成员管理
	member.RemoveMember(p.accountId)

	// 断开WS连接
	if p.client != nil {
		p.client.Disconnect()
	}

	return nil
}

// GetAccountId 获取账号ID
func (p *Plugin) GetAccountId() string {
	return p.accountId
}

// GetState 获取状态
func (p *Plugin) GetState() string {
	if p.client == nil {
		return "disconnected"
	}
	return p.client.GetState()
}

// OnReady 连接就绪
func (p *Plugin) OnReady(data *types.OnReadyData) {
	p.log.Info("WebSocket连接就绪", logger.F("ConnectID", data.ConnectID))

	if p.runtime != nil && p.runtime.onConnected != nil {
		p.runtime.onConnected()
	}
}

// OnDispatch 消息推送
func (p *Plugin) OnDispatch(msg *types.WsMessageMsg) {
	p.log.Debug("收到推送", logger.F("msgId", msg.Payload.MessageId))

	// 调用消息处理回调
	if p.runtime != nil && p.runtime.onMessage != nil {
		p.runtime.onMessage(msg)
	}
}

// OnStateChange 状态变化
func (p *Plugin) OnStateChange(state string) {
	p.log.Info("WebSocket状态变化", logger.F("state", state))

	if state == "disconnected" && p.runtime != nil && p.runtime.onDisconnected != nil {
		p.runtime.onDisconnected()
	}
}

// OnError 错误
func (p *Plugin) OnError(err error) {
	p.log.Error("WebSocket错误", logger.F("error", err.Error()))
}

// OnClose 关闭
func (p *Plugin) OnClose(code int, reason string) {
	p.log.Info("WebSocket关闭",
		logger.F("code", code),
		logger.F("reason", reason),
	)
}

// OnKickout 被踢
func (p *Plugin) OnKickout(code int, reason string) {
	p.log.Warn("被踢下线",
		logger.F("code", code),
		logger.F("reason", reason),
	)
}

// OnAuthFailed 认证失败
func (p *Plugin) OnAuthFailed(code int) (*types.WsAuth, error) {
	p.log.Warn("认证失败，尝试刷新Token", logger.F("code", code))
	return nil, nil
}

// SendMessage 发送消息
func (p *Plugin) SendMessage(to string, text string) (*types.SendResult, error) {
	if p.client == nil || p.client.GetState() != "connected" {
		p.log.Warn("发送消息失败：未连接")
		return &types.SendResult{
			Ok:    false,
			Error: fmt.Errorf("not connected"),
		}, nil
	}

	// 解析目标
	targetId := parseTarget(to)

	p.log.Debug("发送消息",
		logger.F("target", targetId),
		logger.F("text", text),
	)

	messageId := fmt.Sprintf("msg_%d", time.Now().UnixNano())

	err := p.client.SendMessage(targetId, text, messageId, 0, true)
	if err != nil {
		p.log.Error("发送消息失败", logger.F("error", err.Error()))
		return &types.SendResult{
			Ok:        false,
			MessageID: messageId,
			Error:     err,
		}, nil
	}

	return &types.SendResult{Ok: true, MessageID: messageId}, nil
}

// parseTarget 解析目标
func parseTarget(to string) string {
	toUserId := to
	if len(toUserId) > 5 && toUserId[:5] == "user:" {
		toUserId = toUserId[5:]
	}
	return toUserId
}

// SetRuntime 设置运行时
func (p *Plugin) SetRuntime(runtime *Runtime) {
	p.runtime = runtime
}

// SetOnMessage 设置消息处理回调
func (p *Plugin) SetOnMessage(fn func(msg *types.WsMessageMsg)) {
	if p.runtime == nil {
		p.runtime = &Runtime{}
	}
	p.runtime.onMessage = fn
}

// SetOnConnected 设置连接成功回调
func (p *Plugin) SetOnConnected(fn func()) {
	if p.runtime == nil {
		p.runtime = &Runtime{}
	}
	p.runtime.onConnected = fn
}

// SetOnDisconnected 设置断开连接回调
func (p *Plugin) SetOnDisconnected(fn func()) {
	if p.runtime == nil {
		p.runtime = &Runtime{}
	}
	p.runtime.onDisconnected = fn
}

// GetMember 获取成员管理
func (p *Plugin) GetMember() member.MemberInterface {
	return member.NewManager()
}

// GetTokenManager 获取 Token 管理器
func (p *Plugin) GetTokenManager() *account.TokenManager {
	return p.tokenManager
}

// PluginManager 插件管理器
type PluginManager struct {
	mu      sync.RWMutex
	plugins map[string]*Plugin
	log     *logger.Logger
}

// NewPluginManager 创建插件管理器
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]*Plugin),
		log:     logger.New("plugin-manager"),
	}
}

// CreatePlugin 创建插件
func (m *PluginManager) CreatePlugin(accountId string, account *types.ResolvedAccount, cfg *types.Config) *Plugin {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 停止已有插件
	if existing, ok := m.plugins[accountId]; ok {
		existing.Stop()
	}

	plugin := New(accountId, account, cfg)
	m.plugins[accountId] = plugin

	return plugin
}

// GetPlugin 获取插件
func (m *PluginManager) GetPlugin(accountId string) *Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.plugins[accountId]
}

// StopPlugin 停止插件
func (m *PluginManager) StopPlugin(accountId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, ok := m.plugins[accountId]
	if !ok {
		return nil
	}

	err := plugin.Stop()
	delete(m.plugins, accountId)
	return err
}

// StopAll 停止所有插件
func (m *PluginManager) StopAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, plugin := range m.plugins {
		plugin.Stop()
	}

	m.plugins = make(map[string]*Plugin)
	return nil
}

// 全局插件管理器
var (
	globalPluginManager *PluginManager
	pluginManagerOnce   sync.Once
)

// GetPluginManager 获取全局插件管理器
func GetPluginManager() *PluginManager {
	pluginManagerOnce.Do(func() {
		globalPluginManager = NewPluginManager()
	})
	return globalPluginManager
}

// CreateAndStart 创建并启动插件
func CreateAndStart(accountId string, account *types.ResolvedAccount, cfg *types.Config) (*Plugin, error) {
	manager := GetPluginManager()
	plugin := manager.CreatePlugin(accountId, account, cfg)

	if err := plugin.Start(); err != nil {
		return nil, err
	}

	return plugin, nil
}

// StopAndRemove 停止并移除插件
func StopAndRemove(accountId string) error {
	return GetPluginManager().StopPlugin(accountId)
}

// GetPluginByAccountId 根据账号ID获取插件
func GetPluginByAccountId(accountId string) *Plugin {
	return GetPluginManager().GetPlugin(accountId)
}

// RunWithContext 运行直到上下文取消
func RunWithContext(ctx context.Context, accountId string, account *types.ResolvedAccount, cfg *types.Config) error {
	plugin, err := CreateAndStart(accountId, account, cfg)
	if err != nil {
		return err
	}

	// 设置消息处理
	plugin.SetOnMessage(func(msg *types.WsMessageMsg) {
		// TODO: 处理消息逻辑
	})

	// 运行直到取消
	<-ctx.Done()

	return plugin.Stop()
}

// maskString 遮蔽字符串
func maskString(s string) string {
	if len(s) <= 6 {
		return "***"
	}
	return s[:3] + "..." + s[len(s)-3:]
}
