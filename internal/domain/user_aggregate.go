package domain

import (
	"fmt"
	"time"
)

// UserAggregate 用户聚合根
// 聚合根是一致性边界，负责协调聚合内的所有对象
type UserAggregate struct {
	user    *User
	events  *EventRecorder
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
		user:   user,
		events: NewEventRecorder(),
	}, nil
}

// RebuildUserAggregate 从已有用户重建聚合根
func RebuildUserAggregate(user *User) *UserAggregate {
	return &UserAggregate{
		user:   user,
		events: NewEventRecorder(),
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
	oldStatus := a.user.Status()

	// 执行状态变更
	if err := a.user.ChangeStatus(newStatus); err != nil {
		return err
	}

	// 记录领域事件
	a.events.AddEvent(NewUserStatusChangedEvent(
		a.user.ID(),
		a.user.Email().String(),
		a.user.Name().String(),
		oldStatus,
		newStatus,
	))

	return nil
}

// UpdateProfile 更新用户资料
func (a *UserAggregate) UpdateProfile(name Name, email Email) error {
	oldName := a.user.Name().String()
	oldEmail := a.user.Email().String()

	if err := a.user.UpdateProfile(name, email); err != nil {
		return err
	}

	// 记录领域事件
	a.events.AddEvent(NewUserUpdatedEvent(
		a.user.ID(),
		a.user.Email().String(),
		a.user.Name().String(),
	))

	return nil
}

// ChangePassword 更改密码
func (a *UserAggregate) ChangePassword(hashedPassword HashedPassword) error {
	// 密码变更不记录详细事件（出于安全考虑），但可以记录审计日志
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

// Events 返回聚合根中发生的所有领域事件
func (a *UserAggregate) Events() []DomainEvent {
	return a.events.Events()
}

// ClearEvents 清除已发布的事件
func (a *UserAggregate) ClearEvents() {
	a.events.ClearEvents()
}

// HasEvents 检查是否有未发布的事件
func (a *UserAggregate) HasEvents() bool {
	return a.events.HasEvents()
}