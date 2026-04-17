package service

import (
	"context"

	"example.com/classic/internal/domain"
	"example.com/classic/internal/service/dto"
	"example.com/classic/pkg/errors"
	"example.com/classic/pkg/logger"
	"example.com/classic/pkg/tracer"
)

// UserService defines the user service interface
type UserService interface {
	Register(ctx context.Context, params *dto.RegisterParams) (*domain.User, error)
	GetByID(ctx context.Context, id int) (*domain.User, error)
	Update(ctx context.Context, id int, params *dto.UpdateParams) (*domain.User, error)
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, query *dto.UserQueryParams) ([]*domain.User, int64, error)
	ChangeStatus(ctx context.Context, id int, status domain.Status) error
}

// userService user service implementation (application service layer)
type userService struct {
	userRepo       domain.UserRepository
	userFactory    domain.UserFactory
	txManager      domain.TransactionManager
	eventPublisher domain.EventPublisher
	log            logger.Logger
}

// NewUserService creates user service instance
func NewUserService(
	userRepo domain.UserRepository,
	userFactory domain.UserFactory,
	txManager domain.TransactionManager,
	eventPublisher domain.EventPublisher,
	log logger.Logger,
) UserService {
	return &userService{
		userRepo:       userRepo,
		userFactory:    userFactory,
		txManager:      txManager,
		eventPublisher: eventPublisher,
		log:            log,
	}
}

// Register user registration
func (s *userService) Register(ctx context.Context, params *dto.RegisterParams) (*domain.User, error) {
	// 创建 Service 层 span
	span, ctx := tracer.ServiceSpan(ctx, s.log, "Register")
	defer span.End()

	s.log.Info(ctx, "用户注册开始",
		logger.String("email", params.Email),
		logger.String("name", params.Name))

	var user *domain.User

	// Execute within transaction
	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// 1. Use domain factory to create user aggregate (business logic in domain layer)
		factorySpan, txCtx := tracer.StartSpan(txCtx, s.log, "domain:CreateNewUser")
		aggregate, err := s.userFactory.CreateNewUser(params.Name, params.Email, params.Password)
		if err != nil {
			factorySpan.EndWithError(err)
			s.log.Warn(ctx, "创建用户聚合失败", logger.Err(err))
			return errors.New(errors.ErrCodeInvalidParam, err.Error())
		}
		factorySpan.End()

		// 2. Check email uniqueness (application service coordination)
		checkSpan, txCtx := tracer.DBSpan(txCtx, s.log, "SELECT COUNT(*) FROM users WHERE email=?")
		exists, err := s.userRepo.ExistsByEmail(txCtx, params.Email)
		if err != nil {
			checkSpan.EndWithError(err)
			return err
		}
		checkSpan.End()

		if exists {
			s.log.Warn(ctx, "邮箱已存在", logger.String("email", params.Email))
			return errors.ErrUserAlreadyExists
		}

		// 3. Persist aggregate
		saveSpan, txCtx := tracer.DBSpan(txCtx, s.log, "INSERT INTO users")
		if err := s.userRepo.Save(txCtx, aggregate); err != nil {
			saveSpan.EndWithError(err)
			return err
		}
		saveSpan.End()

		user = aggregate.User()

		// 4. Publish domain events (decoupled business logic)
		// Event handlers (e.g. welcome email) are triggered via EventPublisher
		if aggregate.HasEvents() {
			eventSpan, _ := tracer.StartSpan(txCtx, s.log, "event:PublishBatch")
			if err := s.eventPublisher.PublishBatch(aggregate.Events()); err != nil {
				eventSpan.EndWithError(err)
				s.log.Warn(ctx, "failed to publish domain events", logger.Err(err))
			} else {
				eventSpan.End()
			}
			aggregate.ClearEvents()
		}

		return nil
	})

	if err != nil {
		span.EndWithError(err)
		return nil, err
	}

	s.log.Info(ctx, "用户注册完成",
		logger.Int("user_id", user.ID()),
		logger.String("email", user.Email().String()))

	return user, nil
}

// GetByID 根据ID获取用户
func (s *userService) GetByID(ctx context.Context, id int) (*domain.User, error) {
	// 创建 Service 层 span
	span, ctx := tracer.ServiceSpan(ctx, s.log, "GetByID")
	defer span.End()

	s.log.Debug(ctx, "根据ID获取用户", logger.Int("user_id", id))

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		span.EndWithError(err)
		return nil, err
	}

	// 清除敏感信息
	user.ClearSensitiveData()
	return user, nil
}

// Update updates user
func (s *userService) Update(ctx context.Context, id int, params *dto.UpdateParams) (*domain.User, error) {
	// 创建 Service 层 span
	span, ctx := tracer.ServiceSpan(ctx, s.log, "Update")
	defer span.End()

	s.log.Info(ctx, "更新用户",
		logger.Int("user_id", id),
		logger.Bool("has_name", params.Name != nil),
		logger.Bool("has_email", params.Email != nil))

	// 1. 获取聚合根
	aggregate, err := s.userRepo.GetAggregateByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. 更新资料（业务逻辑在领域对象中）
	if params.Name != nil || params.Email != nil {
		var name domain.Name
		var email domain.Email

		if params.Name != nil {
			nameVO, err := domain.NewName(*params.Name)
			if err != nil {
				return nil, errors.New(errors.ErrCodeInvalidParam, err.Error())
			}
			name = *nameVO
		} else {
			name = aggregate.User().Name()
		}

		if params.Email != nil {
			emailVO, err := domain.NewEmail(*params.Email)
			if err != nil {
				return nil, errors.New(errors.ErrCodeInvalidParam, err.Error())
			}
			email = *emailVO

			// 检查邮箱唯一性
			exists, err := s.userRepo.ExistsByEmail(ctx, email.String())
			if err != nil {
				return nil, err
			}
			if exists && email.String() != aggregate.User().Email().String() {
				return nil, errors.ErrUserAlreadyExists
			}
		} else {
			email = aggregate.User().Email()
		}

		if err := aggregate.UpdateProfile(name, email); err != nil {
			return nil, errors.WrapInternalError(err, "failed to update profile")
		}
	}

	// 3. 更新状态（如果提供）
	if params.Status != nil {
		if err := aggregate.ChangeStatus(*params.Status); err != nil {
			return nil, errors.New(errors.ErrCodeInvalidParam, err.Error())
		}
	}

	// 4. 持久化
	if err := s.userRepo.Save(ctx, aggregate); err != nil {
		return nil, err
	}

	// 5. 发布领域事件
	if aggregate.HasEvents() {
		if err := s.eventPublisher.PublishBatch(aggregate.Events()); err != nil {
			s.log.Warn(ctx, "failed to publish domain events", logger.Err(err))
		}
		aggregate.ClearEvents()
	}

	user := aggregate.User()
	user.ClearSensitiveData()

	s.log.Info(ctx, "user updated successfully", logger.F("user_id", id))
	return user, nil
}

// Delete deletes user
func (s *userService) Delete(ctx context.Context, id int) error {
	// 创建 Service 层 span
	span, ctx := tracer.ServiceSpan(ctx, s.log, "Delete")
	defer span.End()

	s.log.Info(ctx, "删除用户", logger.Int("user_id", id))

	// 1. 获取聚合根
	aggregate, err := s.userRepo.GetAggregateByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. 检查是否可以删除（业务规则）
	if err := aggregate.CanBeDeleted(); err != nil {
		return errors.New(errors.ErrCodeInvalidParam, err.Error())
	}

	// 3. 执行删除
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.log.Info(ctx, "user deleted successfully", logger.F("user_id", id))
	return nil
}

// List queries user list
func (s *userService) List(ctx context.Context, query *dto.UserQueryParams) ([]*domain.User, int64, error) {
	// 创建 Service 层 span
	span, ctx := tracer.ServiceSpan(ctx, s.log, "List")
	defer span.End()

	s.log.Debug(ctx, "查询用户列表",
		logger.Int("page", query.Page),
		logger.Int("page_size", query.PageSize))

	// Validate and normalize query params
	s.normalizeQuery(query)

	// Convert to domain params
	params := domain.UserListParams{
		ID:       query.ID,
		Name:     query.Name,
		Email:    query.Email,
		Status:   query.Status,
		Page:     query.Page,
		PageSize: query.PageSize,
	}

	// Query user list
	users, total, err := s.userRepo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	// 清除敏感信息
	for _, user := range users {
		user.ClearSensitiveData()
	}

	s.log.Debug(ctx, "users listed successfully",
		logger.F("total", total),
		logger.F("count", len(users)))
	return users, total, nil
}

// ChangeStatus changes user status
func (s *userService) ChangeStatus(ctx context.Context, id int, status domain.Status) error {
	// 创建 Service 层 span
	span, ctx := tracer.ServiceSpan(ctx, s.log, "ChangeStatus")
	defer span.End()

	s.log.Info(ctx, "更改用户状态",
		logger.Int("user_id", id),
		logger.String("new_status", string(status)))

	// 1. 获取聚合根
	aggregate, err := s.userRepo.GetAggregateByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. 改变状态（业务逻辑在领域对象中）
	if err := aggregate.ChangeStatus(status); err != nil {
		return errors.New(errors.ErrCodeInvalidParam, err.Error())
	}

	// 3. 持久化
	if err := s.userRepo.Save(ctx, aggregate); err != nil {
		return err
	}

	// 4. 发布领域事件（解耦业务逻辑）
	// Event handlers (e.g. status change notification) are triggered via EventPublisher
	if aggregate.HasEvents() {
		if err := s.eventPublisher.PublishBatch(aggregate.Events()); err != nil {
			s.log.Warn(ctx, "failed to publish domain events", logger.F("error", err))
		}
		aggregate.ClearEvents()
	}

	s.log.Info(ctx, "user status changed successfully",
		logger.F("user_id", id),
		logger.F("status", status))
	return nil
}

// normalizeQuery normalizes query parameters
func (s *userService) normalizeQuery(query *dto.UserQueryParams) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
}