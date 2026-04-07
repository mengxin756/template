package domain

import (
	"fmt"
	"regexp"
	"unicode"
)

// Email 邮箱值对象
type Email struct {
	value string
}

// emailRegex 邮箱验证正则表达式
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// NewEmail 创建邮箱值对象
func NewEmail(email string) (*Email, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if len(email) > 100 {
		return nil, fmt.Errorf("email length cannot exceed 100 characters")
	}
	if !emailRegex.MatchString(email) {
		return nil, fmt.Errorf("invalid email format: %s", email)
	}
	return &Email{value: email}, nil
}

// String 返回邮箱字符串
func (e Email) String() string {
	return e.value
}

// Equals 比较两个邮箱是否相等
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// Password 密码值对象（明文密码）
type Password struct {
	value string
}

// NewPassword 创建密码值对象
func NewPassword(password string) (*Password, error) {
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}
	if len(password) > 100 {
		return nil, fmt.Errorf("password length cannot exceed 100 characters")
	}
	// 可选：添加密码强度检查
	if err := validatePasswordStrength(password); err != nil {
		return nil, err
	}
	return &Password{value: password}, nil
}

// String 返回密码字符串
func (p Password) String() string {
	return p.value
}

// validatePasswordStrength 验证密码强度
func validatePasswordStrength(password string) error {
	var hasUpper, hasLower, hasDigit bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}
	// 放宽要求：至少包含字母和数字
	if !hasLower && !hasUpper {
		return fmt.Errorf("password must contain at least one letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	return nil
}

// Name 姓名值对象
type Name struct {
	value string
}

// NewName 创建姓名值对象
func NewName(name string) (*Name, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if len(name) < 2 {
		return nil, fmt.Errorf("name must be at least 2 characters")
	}
	if len(name) > 50 {
		return nil, fmt.Errorf("name length cannot exceed 50 characters")
	}
	return &Name{value: name}, nil
}

// String 返回姓名字符串
func (n Name) String() string {
	return n.value
}

// Equals 比较两个姓名是否相等
func (n Name) Equals(other Name) bool {
	return n.value == other.value
}

// HashedPassword 哈希密码值对象
type HashedPassword struct {
	value string
}

// NewHashedPassword 从哈希字符串创建哈希密码值对象
func NewHashedPassword(hashedPassword string) (*HashedPassword, error) {
	if hashedPassword == "" {
		return nil, fmt.Errorf("hashed password cannot be empty")
	}
	return &HashedPassword{value: hashedPassword}, nil
}

// String 返回哈希密码字符串
func (h HashedPassword) String() string {
	return h.value
}