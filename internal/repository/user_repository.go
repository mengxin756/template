package repository

import (
	"context"

	"example.com/classic/internal/data/ent"
	"example.com/classic/internal/data/ent/user"
	"example.com/classic/internal/domain"
	"example.com/classic/pkg/errors"
	"example.com/classic/pkg/logger"
)

// userRepository 用户仓储实现
type userRepository struct {
	client *ent.Client
	log    logger.Logger
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(client *ent.Client, log logger.Logger) domain.UserRepository {
	return &userRepository{
		client: client,
		log:    log,
	}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	r.log.Debug(ctx, "creating user", logger.F("email", user.Email))

	// 检查邮箱是否已存在
	exists, err := r.ExistsByEmail(ctx, user.Email)
	if err != nil {
		return errors.WrapInternalError(err, "check email exists failed")
	}
	if exists {
		return errors.ErrUserAlreadyExists
	}

	// 创建 Ent 用户
	entUser, err := r.client.User.Create().
		SetName(user.Name).
		SetEmail(user.Email).
		SetPassword(user.Password).
		SetStatus(string(user.Status)).
		Save(ctx)
	if err != nil {
		r.log.Error(ctx, "create user failed", logger.F("error", err))
		return errors.WrapInternalError(err, "create user failed")
	}

	// 更新领域对象
	user.ID = entUser.ID
	user.CreatedAt = entUser.CreatedAt
	user.UpdatedAt = entUser.UpdatedAt

	r.log.Info(ctx, "user created successfully", logger.F("user_id", user.ID))
	return nil
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	r.log.Debug(ctx, "getting user by id", logger.F("user_id", id))

	entUser, err := r.client.User.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.ErrUserNotFound
		}
		r.log.Error(ctx, "get user by id failed", logger.F("error", err), logger.F("user_id", id))
		return nil, errors.WrapInternalError(err, "get user by id failed")
	}

	return r.entToDomain(entUser), nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.log.Debug(ctx, "getting user by email", logger.F("email", email))

	entUser, err := r.client.User.Query().
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.ErrUserNotFound
		}
		r.log.Error(ctx, "get user by email failed", logger.F("error", err), logger.F("email", email))
		return nil, errors.WrapInternalError(err, "get user by email failed")
	}

	return r.entToDomain(entUser), nil
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	r.log.Debug(ctx, "updating user", logger.F("user_id", user.ID))

	// 检查用户是否存在
	existingUser, err := r.GetByID(ctx, user.ID)
	if err != nil {
		return err
	}

	// 如果邮箱有变化，检查新邮箱是否已存在
	if existingUser.Email != user.Email {
		exists, err := r.ExistsByEmail(ctx, user.Email)
		if err != nil {
			return errors.WrapInternalError(err, "check email exists failed")
		}
		if exists {
			return errors.ErrUserAlreadyExists
		}
	}

	// 更新 Ent 用户
	update := r.client.User.UpdateOneID(user.ID)
	if user.Name != "" {
		update.SetName(user.Name)
	}
	if user.Email != "" {
		update.SetEmail(user.Email)
	}
	if user.Password != "" {
		update.SetPassword(user.Password)
	}
	if user.Status != "" {
		update.SetStatus(string(user.Status))
	}

	entUser, err := update.Save(ctx)
	if err != nil {
		r.log.Error(ctx, "update user failed", logger.F("error", err), logger.F("user_id", user.ID))
		return errors.WrapInternalError(err, "update user failed")
	}

	// 更新领域对象
	user.UpdatedAt = entUser.UpdatedAt

	r.log.Info(ctx, "user updated successfully", logger.F("user_id", user.ID))
	return nil
}

// Delete 删除用户
func (r *userRepository) Delete(ctx context.Context, id int) error {
	r.log.Debug(ctx, "deleting user", logger.F("user_id", id))

	// 检查用户是否存在
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 删除用户
	err = r.client.User.DeleteOneID(id).Exec(ctx)
	if err != nil {
		r.log.Error(ctx, "delete user failed", logger.F("error", err), logger.F("user_id", id))
		return errors.WrapInternalError(err, "delete user failed")
	}

	r.log.Info(ctx, "user deleted successfully", logger.F("user_id", id))
	return nil
}

// List 查询用户列表
func (r *userRepository) List(ctx context.Context, query *domain.UserQuery) ([]*domain.User, int64, error) {
	r.log.Debug(ctx, "listing users", logger.F("query", query))

	// 构建查询
	entQuery := r.client.User.Query()

	// 添加查询条件
	if query.ID != nil {
		entQuery = entQuery.Where(user.IDEQ(*query.ID))
	}
	if query.Name != nil {
		entQuery = entQuery.Where(user.NameContains(*query.Name))
	}
	if query.Email != nil {
		entQuery = entQuery.Where(user.EmailContains(*query.Email))
	}
	if query.Status != nil {
		entQuery = entQuery.Where(user.StatusEQ(string(*query.Status)))
	}

	// 获取总数
	total, err := entQuery.Count(ctx)
	if err != nil {
		r.log.Error(ctx, "count users failed", logger.F("error", err))
		return nil, 0, errors.WrapInternalError(err, "count users failed")
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	entUsers, err := entQuery.
		Limit(query.PageSize).
		Offset(offset).
		Order(ent.Desc(user.FieldID)).
		All(ctx)
	if err != nil {
		r.log.Error(ctx, "list users failed", logger.F("error", err))
		return nil, 0, errors.WrapInternalError(err, "list users failed")
	}

	// 转换为领域对象
	users := make([]*domain.User, len(entUsers))
	for i, entUser := range entUsers {
		users[i] = r.entToDomain(entUser)
	}

	r.log.Debug(ctx, "users listed successfully", logger.F("total", total), logger.F("count", len(users)))
	return users, int64(total), nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exists, err := r.client.User.Query().
		Where(user.EmailEQ(email)).
		Exist(ctx)
	if err != nil {
		r.log.Error(ctx, "check email exists failed", logger.F("error", err), logger.F("email", email))
		return false, errors.WrapInternalError(err, "check email exists failed")
	}
	return exists, nil
}

// entToDomain 将 Ent 用户转换为领域用户
func (r *userRepository) entToDomain(entUser *ent.User) *domain.User {
	status := domain.Status(entUser.Status)
	if !status.IsValid() {
		status = domain.StatusInactive
	}

	return &domain.User{
		ID:        entUser.ID,
		Name:      entUser.Name,
		Email:     entUser.Email,
		Password:  entUser.Password,
		Status:    status,
		CreatedAt: entUser.CreatedAt,
		UpdatedAt: entUser.UpdatedAt,
	}
}

// domainToEnt 将领域用户转换为 Ent 用户（用于更新）
func (r *userRepository) domainToEnt(user *domain.User) *ent.UserUpdateOne {
	update := r.client.User.UpdateOneID(user.ID)
	if user.Name != "" {
		update.SetName(user.Name)
	}
	if user.Email != "" {
		update.SetEmail(user.Email)
	}
	if user.Password != "" {
		update.SetPassword(user.Password)
	}
	if user.Status != "" {
		update.SetStatus(string(user.Status))
	}
	return update
}
