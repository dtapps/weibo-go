package member

import (
	"fmt"
	"sync"

	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/types"
)

// Manager 成员管理器
type Manager struct {
	mu      sync.RWMutex
	members map[string]*types.Member // userID -> Member
	log     *logger.Logger
}

// 全局成员管理器
var (
	globalManagers map[string]*Manager
	globalMu       sync.RWMutex
	managerOnce    sync.Once
)

// GetManager 获取指定账号的成员管理器
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

// NewManager 创建成员管理器
func NewManager() *Manager {
	return &Manager{
		members: make(map[string]*types.Member),
		log:     logger.New("member"),
	}
}

// AddUser 添加成员
func (m *Manager) AddUser(req *types.MemberAddUserRequest) (*types.Member, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	member := &types.Member{
		UserID:   req.UserID,
		Nickname: req.Nickname,
	}
	m.members[req.UserID] = member

	return member, nil
}

// UpdateUser 更新成员
func (m *Manager) UpdateUser(req *types.MemberUpdateUserRequest) (*types.Member, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if member, ok := m.members[req.UserID]; ok {
		member.Nickname = req.Nickname
		return member, nil
	}

	member := &types.Member{
		UserID:   req.UserID,
		Nickname: req.Nickname,
	}
	m.members[req.UserID] = member

	return member, nil
}

// DeleteUser 删除成员
func (m *Manager) DeleteUser(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.members, userID)

	return nil
}

// ListUsers 列出所有成员
func (m *Manager) ListUsers(req *types.MemberListUsersRequest) *types.MemberListUsersResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()

	members := make([]*types.Member, 0, len(m.members))
	for _, member := range m.members {
		members = append(members, member)
	}

	return &types.MemberListUsersResponse{
		Total:   len(members),
		Members: members,
	}
}

// GetUser 获取成员
func (m *Manager) GetUser(userID string) (*types.Member, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if member, ok := m.members[userID]; ok {
		return member, nil
	}

	return &types.Member{UserID: userID}, fmt.Errorf("member not found: %s", userID)
}

// Clear 清空成员
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.members = make(map[string]*types.Member)
}
