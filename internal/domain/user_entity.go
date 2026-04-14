package domain

import (
	"fmt"
	"time"
)

// User 用户实体（包含业务行为）
type User struct {
	id             int
	name           Name
	email          Email
	hashedPassword HashedPassword
	status         Status
	createdAt      time.Time
	updatedAt      time.Time
}

// NewUser 创建新用户实体（用于从数据库重建）
func NewUser(
	id int,
	name Name,
	email Email,
	hashedPassword HashedPassword,
	status Status,
	createdAt time.Time,
	updatedAt time.Time,
) (*User, error) {
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid status: %s", status)
	}
	return &User{
		id:             id,
		name:           name,
		email:          email,
		hashedPassword: hashedPassword,
		status:         status,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}, nil
}

// ID 获取用户ID
func (u *User) ID() int {
	return u.id
}

// Name 获取用户姓名
func (u *User) Name() Name {
	return u.name
}

// Email 获取用户邮箱
func (u *User) Email() Email {
	return u.email
}

// Status 获取用户状态
func (u *User) Status() Status {
	return u.status
}

// CreatedAt 获取创建时间
func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

// UpdatedAt 获取更新时间
func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// ChangeStatus 改变用户状态（业务行为）
func (u *User) ChangeStatus(newStatus Status) error {
	if !newStatus.IsValid() {
		return fmt.Errorf("invalid status: %s", newStatus)
	}

	// 业务规则：被禁止的用户不能直接激活
	if u.status == StatusBanned && newStatus == StatusActive {
		return fmt.Errorf("cannot directly activate a banned user")
	}

	// 业务规则：状态相同时无需更改
	if u.status == newStatus {
		return fmt.Errorf("user is already in %s status", newStatus)
	}

	u.status = newStatus
	u.updatedAt = time.Now()
	return nil
}

// UpdateProfile 更新用户资料（业务行为）
func (u *User) UpdateProfile(name Name, email Email) error {
	u.name = name
	u.email = email
	u.updatedAt = time.Now()
	return nil
}

// ChangePassword 更改密码（业务行为）
func (u *User) ChangePassword(hashedPassword HashedPassword) error {
	u.hashedPassword = hashedPassword
	u.updatedAt = time.Now()
	return nil
}

// CanBeDeleted 检查用户是否可以被删除（业务规则）
func (u *User) CanBeDeleted() error {
	// 业务规则：活跃用户不能直接删除
	if u.status == StatusActive {
		return fmt.Errorf("active user cannot be deleted, please deactivate first")
	}
	return nil
}

// IsActive 检查用户是否活跃
func (u *User) IsActive() bool {
	return u.status == StatusActive
}

// IsBanned 检查用户是否被禁止
func (u *User) IsBanned() bool {
	return u.status == StatusBanned
}

// SetID 设置用户ID（仅在创建新用户后由仓储调用）
func (u *User) SetID(id int) {
	u.id = id
}

// SetCreatedAt 设置创建时间（由仓储调用）
func (u *User) SetCreatedAt(t time.Time) {
	u.createdAt = t
}

// SetUpdatedAt 设置更新时间（由仓储调用）
func (u *User) SetUpdatedAt(t time.Time) {
	u.updatedAt = t
}

// GetHashedPassword 获取哈希密码（仅在验证时使用）
func (u *User) GetHashedPassword() string {
	return u.hashedPassword.String()
}

// ClearSensitiveData 清除敏感数据
func (u *User) ClearSensitiveData() {
	u.hashedPassword = HashedPassword{}
}

// UserDTO is a data transfer object for user
type UserDTO struct {
	ID        int
	Name      string
	Email     string
	Password  string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ToDTO converts to DTO (for compatibility)
func (u *User) ToDTO() *UserDTO {
	return &UserDTO{
		ID:        u.id,
		Name:      u.name.String(),
		Email:     u.email.String(),
		Password:  "", // password not returned
		Status:    u.status,
		CreatedAt: u.createdAt,
		UpdatedAt: u.updatedAt,
	}
}

// FromDTO 从DTO创建用户实体（用于兼容现有代码）
func FromDTO(dto *UserDTO) (*User, error) {
	name, err := NewName(dto.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid name: %w", err)
	}

	email, err := NewEmail(dto.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	var hashedPassword HashedPassword
	if dto.Password != "" {
		hp, err := NewHashedPassword(dto.Password)
		if err != nil {
			return nil, fmt.Errorf("invalid password: %w", err)
		}
		hashedPassword = *hp
	}

	return NewUser(
		dto.ID,
		*name,
		*email,
		hashedPassword,
		dto.Status,
		dto.CreatedAt,
		dto.UpdatedAt,
	)
}
