package domain

import (
	"fmt"
)

// userFactory 用户工厂实现
type userFactory struct {
	passwordHasher PasswordHasher
}

// NewUserFactory 创建用户工厂
func NewUserFactory(passwordHasher PasswordHasher) UserFactory {
	return &userFactory{
		passwordHasher: passwordHasher,
	}
}

// CreateNewUser 创建新用户聚合根
func (f *userFactory) CreateNewUser(name, email, password string) (*UserAggregate, error) {
	// 1. 创建值对象（验证在值对象内部）
	nameVO, err := NewName(name)
	if err != nil {
		return nil, fmt.Errorf("invalid name: %w", err)
	}

	emailVO, err := NewEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	passwordVO, err := NewPassword(password)
	if err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	// 2. 哈希密码
	hashedPasswordStr, err := f.passwordHasher.Hash(passwordVO.String())
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	hashedPassword, err := NewHashedPassword(hashedPasswordStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create hashed password: %w", err)
	}

	// 3. 创建聚合根
	aggregate, err := NewUserAggregate(*nameVO, *emailVO, *hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create user aggregate: %w", err)
	}

	return aggregate, nil
}