package domain

import (
	"context"
	"time"
)

// User 用户实体
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // 不暴露密码
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Name   *string `json:"name,omitempty" binding:"omitempty,min=2,max=50"`
	Email  *string `json:"email,omitempty" binding:"omitempty,email"`
	Status *Status `json:"status,omitempty" binding:"omitempty,oneof=active inactive banned"`
}

// UserQuery 用户查询条件
type UserQuery struct {
	ID       *int    `json:"id,omitempty"`
	Name     *string `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
	Status   *Status `json:"status,omitempty"`
	Page     int     `json:"page" binding:"min=1"`
	PageSize int     `json:"page_size" binding:"min=1,max=100"`
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
	List(ctx context.Context, query *UserQuery) ([]*User, int64, error)
	
	// ExistsByEmail 检查邮箱是否存在
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// UserService 用户服务接口
type UserService interface {
	// Register 用户注册
	Register(ctx context.Context, req *CreateUserRequest) (*User, error)
	
	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, id int) (*User, error)
	
	// Update 更新用户
	Update(ctx context.Context, id int, req *UpdateUserRequest) (*User, error)
	
	// Delete 删除用户
	Delete(ctx context.Context, id int) error
	
	// List 查询用户列表
	List(ctx context.Context, query *UserQuery) ([]*User, int64, error)
	
	// ChangeStatus 改变用户状态
	ChangeStatus(ctx context.Context, id int, status Status) error
}

// UserHandler HTTP处理器接口
type UserHandler interface {
	// Register 用户注册
	Register(ctx context.Context, req *CreateUserRequest) (*User, error)
	
	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, id int) (*User, error)
	
	// Update 更新用户
	Update(ctx context.Context, id int, req *UpdateUserRequest) (*User, error)
	
	// Delete 删除用户
	Delete(ctx context.Context, id int) error
	
	// List 查询用户列表
	List(ctx context.Context, query *UserQuery) ([]*User, int64, error)
}
