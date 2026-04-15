package repository

import (
	"context"

	"example.com/classic/internal/data/ent"
	"example.com/classic/internal/data/ent/user"
	"example.com/classic/internal/domain"
	"example.com/classic/pkg/errors"
	"example.com/classic/pkg/logger"
	"example.com/classic/pkg/tracer"
)

// userRepository user repository implementation
type userRepository struct {
	client *ent.Client
	log    logger.Logger
}

// NewUserRepository creates user repository instance
func NewUserRepository(client *ent.Client, log logger.Logger) domain.UserRepository {
	return &userRepository{
		client: client,
		log:    log,
	}
}

// getClient returns the appropriate client (transactional or regular)
func (r *userRepository) getClient(ctx context.Context) *ent.Client {
	if tx, ok := domain.TxFromContext(ctx).(*ent.Tx); ok && tx != nil {
		return tx.Client()
	}
	return r.client
}

// Create creates user (backward compatible)
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	// 创建 DB 操作 span
	span, ctx := tracer.DBSpan(ctx, r.log, "INSERT user")
	defer span.End()

	r.log.Debug(ctx, "creating user", logger.F("email", user.Email().String()))

	client := r.getClient(ctx)

	// Check if email already exists
	exists, err := r.ExistsByEmail(ctx, user.Email().String())
	if err != nil {
		span.EndWithError(err)
		return errors.WrapInternalError(err, "check email exists failed")
	}
	if exists {
		return errors.ErrUserAlreadyExists
	}

	// Create Ent user
	entUser, err := client.User.Create().
		SetName(user.Name().String()).
		SetEmail(user.Email().String()).
		SetPassword(user.GetHashedPassword()).
		SetStatus(string(user.Status())).
		Save(ctx)
	if err != nil {
		span.EndWithError(err)
		return errors.WrapInternalError(err, "create user failed")
	}

	// Update domain object
	user.SetID(entUser.ID)
	user.SetCreatedAt(entUser.CreatedAt)
	user.SetUpdatedAt(entUser.UpdatedAt)

	r.log.Info(ctx, "user created successfully", logger.F("user_id", user.ID()))
	return nil
}

// GetByID gets user by ID
func (r *userRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	// 创建 DB 操作 span
	span, ctx := tracer.DBSpan(ctx, r.log, "SELECT user WHERE id=?")
	defer span.End()

	r.log.Debug(ctx, "getting user by id", logger.F("user_id", id))

	client := r.getClient(ctx)
	entUser, err := client.User.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.ErrUserNotFound
		}
		span.EndWithError(err)
		r.log.Error(ctx, "get user by id failed", logger.F("error", err), logger.F("user_id", id))
		return nil, errors.WrapInternalError(err, "get user by id failed")
	}

	return r.entToDomain(entUser)
}

// GetByEmail gets user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.log.Debug(ctx, "getting user by email", logger.F("email", email))

	client := r.getClient(ctx)
	entUser, err := client.User.Query().
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.ErrUserNotFound
		}
		r.log.Error(ctx, "get user by email failed", logger.F("error", err), logger.F("email", email))
		return nil, errors.WrapInternalError(err, "get user by email failed")
	}

	return r.entToDomain(entUser)
}

// Update updates user
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	r.log.Debug(ctx, "updating user", logger.F("user_id", user.ID()))

	client := r.getClient(ctx)

	// Check if user exists
	_, err := r.GetByID(ctx, user.ID())
	if err != nil {
		return err
	}

	// Update Ent user
	entUser, err := client.User.UpdateOneID(user.ID()).
		SetName(user.Name().String()).
		SetEmail(user.Email().String()).
		SetStatus(string(user.Status())).
		Save(ctx)
	if err != nil {
		r.log.Error(ctx, "update user failed", logger.F("error", err), logger.F("user_id", user.ID()))
		return errors.WrapInternalError(err, "update user failed")
	}

	user.SetUpdatedAt(entUser.UpdatedAt)

	r.log.Info(ctx, "user updated successfully", logger.F("user_id", user.ID()))
	return nil
}

// Delete deletes user
func (r *userRepository) Delete(ctx context.Context, id int) error {
	r.log.Debug(ctx, "deleting user", logger.F("user_id", id))

	client := r.getClient(ctx)

	// Check if user exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete user
	err = client.User.DeleteOneID(id).Exec(ctx)
	if err != nil {
		r.log.Error(ctx, "delete user failed", logger.F("error", err), logger.F("user_id", id))
		return errors.WrapInternalError(err, "delete user failed")
	}

	r.log.Info(ctx, "user deleted successfully", logger.F("user_id", id))
	return nil
}

// List queries user list
func (r *userRepository) List(ctx context.Context, params domain.UserListParams) ([]*domain.User, int64, error) {
	r.log.Debug(ctx, "listing users", logger.F("params", params))

	client := r.getClient(ctx)

	// Build query
	entQuery := client.User.Query()

	// Add query conditions
	if params.ID != nil {
		entQuery = entQuery.Where(user.IDEQ(*params.ID))
	}
	if params.Name != nil {
		entQuery = entQuery.Where(user.NameContains(*params.Name))
	}
	if params.Email != nil {
		entQuery = entQuery.Where(user.EmailContains(*params.Email))
	}
	if params.Status != nil {
		entQuery = entQuery.Where(user.StatusEQ(string(*params.Status)))
	}

	// Get total count
	total, err := entQuery.Count(ctx)
	if err != nil {
		r.log.Error(ctx, "count users failed", logger.F("error", err))
		return nil, 0, errors.WrapInternalError(err, "count users failed")
	}

	// Paginated query
	offset := (params.Page - 1) * params.PageSize
	entUsers, err := entQuery.
		Limit(params.PageSize).
		Offset(offset).
		Order(ent.Desc(user.FieldID)).
		All(ctx)
	if err != nil {
		r.log.Error(ctx, "list users failed", logger.F("error", err))
		return nil, 0, errors.WrapInternalError(err, "list users failed")
	}

	// Convert to domain objects
	users := make([]*domain.User, len(entUsers))
	for i, entUser := range entUsers {
		domainUser, err := r.entToDomain(entUser)
		if err != nil {
			return nil, 0, errors.WrapInternalError(err, "convert domain user failed")
		}
		users[i] = domainUser
	}

	r.log.Debug(ctx, "users listed successfully",
		logger.F("total", total),
		logger.F("count", len(users)))
	return users, int64(total), nil
}

// ExistsByEmail checks if email exists
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	client := r.getClient(ctx)
	exists, err := client.User.Query().
		Where(user.EmailEQ(email)).
		Exist(ctx)
	if err != nil {
		r.log.Error(ctx, "check email exists failed", logger.F("error", err), logger.F("email", email))
		return false, errors.WrapInternalError(err, "check email exists failed")
	}
	return exists, nil
}

// Save saves aggregate root
func (r *userRepository) Save(ctx context.Context, aggregate *domain.UserAggregate) error {
	user := aggregate.User()

	if user.ID() == 0 {
		// Create new
		return r.Create(ctx, user)
	}

	// Update existing
	return r.Update(ctx, user)
}

// GetAggregateByID gets aggregate by ID
func (r *userRepository) GetAggregateByID(ctx context.Context, id int) (*domain.UserAggregate, error) {
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Rebuild aggregate from user entity
	aggregate := domain.RebuildUserAggregate(user)
	return aggregate, nil
}

// GetAggregateByEmail gets aggregate by email
func (r *userRepository) GetAggregateByEmail(ctx context.Context, email string) (*domain.UserAggregate, error) {
	user, err := r.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Rebuild aggregate from user entity
	aggregate := domain.RebuildUserAggregate(user)
	return aggregate, nil
}

// entToDomain converts Ent user to domain user
func (r *userRepository) entToDomain(entUser *ent.User) (*domain.User, error) {
	status := domain.Status(entUser.Status)
	if !status.IsValid() {
		status = domain.StatusInactive
	}

	name, err := domain.NewName(entUser.Name)
	if err != nil {
		return nil, errors.WrapInternalError(err, "invalid user name from database")
	}

	email, err := domain.NewEmail(entUser.Email)
	if err != nil {
		return nil, errors.WrapInternalError(err, "invalid user email from database")
	}

	hashedPassword, err := domain.NewHashedPassword(entUser.Password)
	if err != nil {
		return nil, errors.WrapInternalError(err, "invalid user password from database")
	}

	user, err := domain.NewUser(
		entUser.ID,
		*name,
		*email,
		*hashedPassword,
		status,
		entUser.CreatedAt,
		entUser.UpdatedAt,
	)
	if err != nil {
		return nil, errors.WrapInternalError(err, "create user entity failed")
	}

	return user, nil
}