package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"example.com/classic/internal/domain"
	"example.com/classic/pkg/errors"
	"example.com/classic/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// userService 用户服务实现
type userService struct {
	userRepo domain.UserRepository
	log      logger.Logger
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo domain.UserRepository, log logger.Logger) domain.UserService {
	return &userService{
		userRepo: userRepo,
		log:      log,
	}
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	s.log.Info(ctx, "user registration started", logger.F("email", req.Email))
	
	// 验证请求参数
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}
	
	// 检查邮箱是否已存在
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrUserAlreadyExists
	}
	
	// 加密密码
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		s.log.Error(ctx, "password hashing failed", logger.F("error", err))
		return nil, errors.WrapInternalError(err, "password hashing failed")
	}
	
	// 创建用户
	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Status:   domain.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	
	s.log.Info(ctx, "user registration completed", logger.F("user_id", user.ID), logger.F("email", user.Email))
	return user, nil
}

// GetByID 根据ID获取用户
func (s *userService) GetByID(ctx context.Context, id int) (*domain.User, error) {
	s.log.Debug(ctx, "getting user by id", logger.F("user_id", id))
	
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// 清除敏感信息
	user.Password = ""
	return user, nil
}

// Update 更新用户
func (s *userService) Update(ctx context.Context, id int, req *domain.UpdateUserRequest) (*domain.User, error) {
	s.log.Info(ctx, "updating user", logger.F("user_id", id))
	
	// 获取现有用户
	existingUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// 应用更新
	if req.Name != nil {
		existingUser.Name = *req.Name
	}
	if req.Email != nil {
		existingUser.Email = *req.Email
	}
	if req.Status != nil {
		if !req.Status.IsValid() {
			return nil, errors.New(errors.ErrCodeInvalidParam, "invalid status")
		}
		existingUser.Status = *req.Status
	}
	
	existingUser.UpdatedAt = time.Now()
	
	// 保存更新
	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		return nil, err
	}
	
	// 清除敏感信息
	existingUser.Password = ""
	
	s.log.Info(ctx, "user updated successfully", logger.F("user_id", id))
	return existingUser, nil
}

// Delete 删除用户
func (s *userService) Delete(ctx context.Context, id int) error {
	s.log.Info(ctx, "deleting user", logger.F("user_id", id))
	
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	// 执行删除
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}
	
	s.log.Info(ctx, "user deleted successfully", logger.F("user_id", id))
	return nil
}

// List 查询用户列表
func (s *userService) List(ctx context.Context, query *domain.UserQuery) ([]*domain.User, int64, error) {
	s.log.Debug(ctx, "listing users", logger.F("query", query))
	
	// 验证查询参数
	if err := s.validateQuery(query); err != nil {
		return nil, 0, err
	}
	
	// 查询用户列表
	users, total, err := s.userRepo.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	
	// 清除敏感信息
	for _, user := range users {
		user.Password = ""
	}
	
	s.log.Debug(ctx, "users listed successfully", logger.F("total", total), logger.F("count", len(users)))
	return users, total, nil
}

// ChangeStatus 改变用户状态
func (s *userService) ChangeStatus(ctx context.Context, id int, status domain.Status) error {
	s.log.Info(ctx, "changing user status", logger.F("user_id", id), logger.F("status", status))
	
	// 验证状态
	if !status.IsValid() {
		return errors.New(errors.ErrCodeInvalidParam, "invalid status")
	}
	
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	// 更新状态
	user.Status = status
	user.UpdatedAt = time.Now()
	
	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}
	
	s.log.Info(ctx, "user status changed successfully", logger.F("user_id", id), logger.F("status", status))
	return nil
}

// validateCreateRequest 验证创建用户请求
func (s *userService) validateCreateRequest(req *domain.CreateUserRequest) error {
	if req.Name == "" {
		return errors.New(errors.ErrCodeInvalidParam, "name is required")
	}
	if len(req.Name) < 2 || len(req.Name) > 50 {
		return errors.New(errors.ErrCodeInvalidParam, "name length must be between 2 and 50")
	}
	
	if req.Email == "" {
		return errors.New(errors.ErrCodeInvalidParam, "email is required")
	}
	if !s.isValidEmail(req.Email) {
		return errors.New(errors.ErrCodeInvalidParam, "invalid email format")
	}
	
	if req.Password == "" {
		return errors.New(errors.ErrCodeInvalidParam, "password is required")
	}
	if len(req.Password) < 6 || len(req.Password) > 100 {
		return errors.New(errors.ErrCodeInvalidParam, "password length must be between 6 and 100")
	}
	
	return nil
}

// validateQuery 验证查询参数
func (s *userService) validateQuery(query *domain.UserQuery) error {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	return nil
}

// isValidEmail 简单的邮箱格式验证
func (s *userService) isValidEmail(email string) bool {
	// 这里可以添加更复杂的邮箱验证逻辑
	return len(email) > 0 && len(email) <= 100
}

// hashPassword 加密密码
func (s *userService) hashPassword(password string) (string, error) {
	// 生成随机盐值
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	
	// 使用 bcrypt 加密
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	
	return hex.EncodeToString(hashedBytes), nil
}

// verifyPassword 验证密码
func (s *userService) verifyPassword(hashedPassword, password string) error {
	hashedBytes, err := hex.DecodeString(hashedPassword)
	if err != nil {
		return err
	}
	
	return bcrypt.CompareHashAndPassword(hashedBytes, []byte(password))
}
