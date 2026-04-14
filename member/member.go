package member

// MemberInterface 成员接口
type MemberInterface interface {
	GetMember(userId string) (*Member, error)
}

// Manager 成员管理器
type Manager struct{}

// NewManager 创建成员管理器
func NewManager() *Manager {
	return &Manager{}
}

// RemoveMember 移除成员
func RemoveMember(accountId string) {}

// GetMember 获取成员
func (m *Manager) GetMember(userId string) (*Member, error) {
	return &Member{ID: userId}, nil
}

// Member 成员
type Member struct {
	ID       string
	Nickname string
}
