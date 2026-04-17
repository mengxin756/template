package hashing

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptPasswordHasher bcrypt密码哈希器实现
type BcryptPasswordHasher struct {
	cost int
}

// NewBcryptPasswordHasher 创建bcrypt密码哈希器
func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{
		cost: bcrypt.DefaultCost,
	}
}

// Hash 哈希密码
func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash failed: %w", err)
	}
	return string(hashedBytes), nil
}

// Verify 验证密码
func (h *BcryptPasswordHasher) Verify(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}
	return nil
}
