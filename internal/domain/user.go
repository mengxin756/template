package domain

import (
	"context"
)

// Status 用户状态
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusBanned   Status = "banned"
)

// IsValid 验证状态是否有效
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusBanned:
		return true
	default:
		return false
	}
}

// String 返回状态字符串
func (s Status) String() string {
	return string(s)
}

// UserRepository 用户仓储接口
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *User) error

	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, id int) (*User, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Update 更新用户
	Update(ctx context.Context, user *User) error

	// Delete 删除用户
	Delete(ctx context.Context, id int) error

	// List 查询用户列表
	List(ctx context.Context, params UserListParams) ([]*User, int64, error)

	// ExistsByEmail 检查邮箱是否存在
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// Save 保存聚合根
	Save(ctx context.Context, aggregate *UserAggregate) error

	// GetAggregateByID 根据ID获取聚合根
	GetAggregateByID(ctx context.Context, id int) (*UserAggregate, error)

	// GetAggregateByEmail 根据邮箱获取聚合根
	GetAggregateByEmail(ctx context.Context, email string) (*UserAggregate, error)
}

// UserListParams 用户列表查询参数
type UserListParams struct {
	ID       *int
	Name     *string
	Email    *string
	Status   *Status
	Page     int
	PageSize int
}

// PasswordHasher 密码哈希器接口（领域服务）
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hashedPassword, password string) error
}

// UserFactory 用户工厂接口（领域服务）
type UserFactory interface {
	// CreateNewUser 创建新用户聚合根
	CreateNewUser(name, email, password string) (*UserAggregate, error)
}