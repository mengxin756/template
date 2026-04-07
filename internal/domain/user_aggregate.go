package domain

import (
	"fmt"
	"time"
)

// UserAggregate 用户聚合根
// 聚合根是一致性边界，负责协调聚合内的所有对象
type UserAggregate struct {
	user *User
	// 未来可以扩展：
	// profile   *Profile    // 用户资料实体
	// settings  *Settings   // 用户设置实体
	// addresses []*Address  // 用户地址列表
}

// NewUserAggregate 创建新的用户聚合根
func NewUserAggregate(
	name Name,
	email Email,
	hashedPassword HashedPassword,
) (*UserAggregate, error) {
	now := time.Now()

	user, err := NewUser(
		0, // ID 由数据库生成
		name,
		email,
		hashedPassword,
		StatusActive, // 新用户默认活跃
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &UserAggregate{
		user: user,
	}, nil
}

// RebuildUserAggregate 从已有用户重建聚合根
func RebuildUserAggregate(user *User) *UserAggregate {
	return &UserAggregate{
		user: user,
	}
}

// User 获取聚合根中的用户实体
func (a *UserAggregate) User() *User {
	return a.user
}

// ID 获取用户ID
func (a *UserAggregate) ID() int {
	return a.user.ID()
}

// ChangeStatus 改变用户状态（聚合根协调）
func (a *UserAggregate) ChangeStatus(newStatus Status) error {
	// 可以在这里添加跨实体的业务规则
	// 例如：如果用户有未完成的订单，不能禁用

	return a.user.ChangeStatus(newStatus)
}

// UpdateProfile 更新用户资料
func (a *UserAggregate) UpdateProfile(name Name, email Email) error {
	return a.user.UpdateProfile(name, email)
}

// ChangePassword 更改密码
func (a *UserAggregate) ChangePassword(hashedPassword HashedPassword) error {
	return a.user.ChangePassword(hashedPassword)
}

// Deactivate 停用用户
func (a *UserAggregate) Deactivate() error {
	return a.user.ChangeStatus(StatusInactive)
}

// Ban 封禁用户
func (a *UserAggregate) Ban() error {
	return a.user.ChangeStatus(StatusBanned)
}

// Activate 激活用户
func (a *UserAggregate) Activate() error {
	// 业务规则：被禁止的用户不能直接激活
	if a.user.IsBanned() {
		return fmt.Errorf("cannot activate a banned user")
	}
	return a.user.ChangeStatus(StatusActive)
}

// CanBeDeleted 检查是否可以删除
func (a *UserAggregate) CanBeDeleted() error {
	return a.user.CanBeDeleted()
}

// IsActive 检查是否活跃
func (a *UserAggregate) IsActive() bool {
	return a.user.IsActive()
}

// IsBanned 检查是否被封禁
func (a *UserAggregate) IsBanned() bool {
	return a.user.IsBanned()
}