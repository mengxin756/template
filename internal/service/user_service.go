package service

import (
	"context"

	"example.com/classic/internal/domain"
	"example.com/classic/internal/job/asynq"
	"example.com/classic/pkg/errors"
	"example.com/classic/pkg/logger"
)

// userService 用户服务实现（应用服务层）
type userService struct {
	userRepo       domain.UserRepository
	userFactory    domain.UserFactory
	taskQueue      *asynq.Queue
	eventPublisher domain.EventPublisher
	log            logger.Logger
}

// NewUserService 创建用户服务实例
func NewUserService(
	userRepo domain.UserRepository,
	userFactory domain.UserFactory,
	taskQueue *asynq.Queue,
	eventPublisher domain.EventPublisher,
	log logger.Logger,
) domain.UserService {
	return &userService{
		userRepo:       userRepo,
		userFactory:    userFactory,
		taskQueue:      taskQueue,
		eventPublisher: eventPublisher,
		log:            log,
	}
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	s.log.Info(ctx, "user registration started", logger.F("email", req.Email))

	// 1. 使用领域工厂创建用户聚合根（业务逻辑在领域层）
	aggregate, err := s.userFactory.CreateNewUser(req.Name, req.Email, req.Password)
	if err != nil {
		s.log.Warn(ctx, "failed to create user aggregate", logger.F("error", err))
		return nil, errors.New(errors.ErrCodeInvalidParam, err.Error())
	}

	// 2. 检查邮箱唯一性（应用服务协调）
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrUserAlreadyExists
	}

	// 3. 持久化聚合根
	if err := s.userRepo.Save(ctx, aggregate); err != nil {
		return nil, err
	}

	user := aggregate.User()

	// 4. 发布领域事件（解耦业务逻辑）
	if aggregate.HasEvents() {
		if err := s.eventPublisher.PublishBatch(aggregate.Events()); err != nil {
			s.log.Warn(ctx, "failed to publish domain events", logger.F("error", err))
			// 不阻塞主流程，只记录警告
		}
		// 清除已发布的事件
		aggregate.ClearEvents()
	}

	// 5. 发送欢迎邮件任务（应用服务协调）
	if s.taskQueue != nil {
		s.enqueueWelcomeEmail(ctx, user.ID(), user.Email().String(), user.Name().String())
	}

	s.log.Info(ctx, "user registration completed",
		logger.F("user_id", user.ID()),
		logger.F("email", user.Email().String()))

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
	user.ClearSensitiveData()
	return user, nil
}

// Update 更新用户
func (s *userService) Update(ctx context.Context, id int, req *domain.UpdateUserRequest) (*domain.User, error) {
	s.log.Info(ctx, "updating user", logger.F("user_id", id))

	// 1. 获取聚合根
	aggregate, err := s.userRepo.GetAggregateByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. 更新资料（业务逻辑在领域对象中）
	if req.Name != nil || req.Email != nil {
		var name domain.Name
		var email domain.Email

		if req.Name != nil {
			nameVO, err := domain.NewName(*req.Name)
			if err != nil {
				return nil, errors.New(errors.ErrCodeInvalidParam, err.Error())
			}
			name = *nameVO
		} else {
			name = aggregate.User().Name()
		}

		if req.Email != nil {
			emailVO, err := domain.NewEmail(*req.Email)
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
	if req.Status != nil {
		if err := aggregate.ChangeStatus(*req.Status); err != nil {
			return nil, errors.New(errors.ErrCodeInvalidParam, err.Error())
		}
	}

	// 4. 持久化
	if err := s.userRepo.Save(ctx, aggregate); err != nil {
		return nil, err
	}

	user := aggregate.User()
	user.ClearSensitiveData()

	s.log.Info(ctx, "user updated successfully", logger.F("user_id", id))
	return user, nil
}

// Delete 删除用户
func (s *userService) Delete(ctx context.Context, id int) error {
	s.log.Info(ctx, "deleting user", logger.F("user_id", id))

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

// List 查询用户列表
func (s *userService) List(ctx context.Context, query *domain.UserQuery) ([]*domain.User, int64, error) {
	s.log.Debug(ctx, "listing users", logger.F("query", query))

	// 验证查询参数
	s.normalizeQuery(query)

	// 查询用户列表
	users, total, err := s.userRepo.List(ctx, query)
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

// ChangeStatus 改变用户状态
func (s *userService) ChangeStatus(ctx context.Context, id int, status domain.Status) error {
	s.log.Info(ctx, "changing user status", logger.F("user_id", id))

	// 1. 获取聚合根
	aggregate, err := s.userRepo.GetAggregateByID(ctx, id)
	if err != nil {
		return err
	}

	oldStatus := aggregate.User().Status()

	// 2. 改变状态（业务逻辑在领域对象中）
	if err := aggregate.ChangeStatus(status); err != nil {
		return errors.New(errors.ErrCodeInvalidParam, err.Error())
	}

	// 3. 持久化
	if err := s.userRepo.Save(ctx, aggregate); err != nil {
		return err
	}

	// 4. 发布领域事件（解耦业务逻辑）
	if aggregate.HasEvents() {
		if err := s.eventPublisher.PublishBatch(aggregate.Events()); err != nil {
			s.log.Warn(ctx, "failed to publish domain events", logger.F("error", err))
			// 不阻塞主流程，只记录警告
		}
		// 清除已发布的事件
		aggregate.ClearEvents()
	}

	// 5. 发送状态变更通知任务（应用服务协调）
	if s.taskQueue != nil && oldStatus != status {
		s.enqueueStatusChangeNotification(ctx,
			aggregate.User().ID(),
			aggregate.User().Email().String(),
			aggregate.User().Name().String(),
			oldStatus.String(),
			status.String())
	}

	s.log.Info(ctx, "user status changed successfully",
		logger.F("user_id", id),
		logger.F("status", status))
	return nil
}

// normalizeQuery 规范化查询参数
func (s *userService) normalizeQuery(query *domain.UserQuery) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
}

// enqueueWelcomeEmail 入队欢迎邮件任务
func (s *userService) enqueueWelcomeEmail(ctx context.Context, userID int, email, name string) {
	task := asynq.NewWelcomeEmailTask(userID, email, name)
	const delaySeconds = 10
	if err := s.taskQueue.EnqueueDelay(delaySeconds*1000000000, task); err != nil {
		s.log.Warn(ctx, "failed to enqueue welcome email task",
			logger.F("error", err),
			logger.F("user_id", userID))
	} else {
		s.log.Debug(ctx, "welcome email task enqueued",
			logger.F("user_id", userID),
			logger.F("email", email))
	}
}

// enqueueStatusChangeNotification 入队状态变更通知任务
func (s *userService) enqueueStatusChangeNotification(ctx context.Context, userID int, email, name, oldStatus, newStatus string) {
	task := asynq.NewStatusChangeNotificationTask(userID, email, name, oldStatus, newStatus, "system")
	if err := s.taskQueue.Enqueue(task); err != nil {
		s.log.Warn(ctx, "failed to enqueue status change notification task",
			logger.F("error", err),
			logger.F("user_id", userID))
	} else {
		s.log.Debug(ctx, "status change notification task enqueued",
			logger.F("user_id", userID),
			logger.F("old_status", oldStatus),
			logger.F("new_status", newStatus))
	}
}