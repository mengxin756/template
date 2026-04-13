package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"example.com/classic/internal/data/db"
	"example.com/classic/internal/domain"
	"example.com/classic/pkg/errors"
	"example.com/classic/pkg/logger"
)

// userRepositorySQLC implements UserRepository using sqlc
type userRepositorySQLC struct {
	queries *db.Queries
	log     logger.Logger
}

// NewUserRepositorySQLC creates a new user repository using sqlc
func NewUserRepositorySQLC(dbtx db.DBTX, log logger.Logger) domain.UserRepository {
	return &userRepositorySQLC{
		queries: db.New(dbtx),
		log:     log,
	}
}

// getQueries returns the appropriate queries (transactional or regular)
func (r *userRepositorySQLC) getQueries(ctx context.Context) *db.Queries {
	if tx, ok := domain.TxFromContext(ctx).(*sql.Tx); ok && tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

// Create creates a new user
func (r *userRepositorySQLC) Create(ctx context.Context, user *domain.User) error {
	r.log.Debug(ctx, "creating user", logger.F("email", user.Email().String()))

	queries := r.getQueries(ctx)

	// Check if email already exists
	exists, err := queries.ExistsByEmail(ctx, user.Email().String())
	if err != nil {
		return errors.WrapInternalError(err, "check email exists failed")
	}
	if exists {
		return errors.ErrUserAlreadyExists
	}

	// Create user
	now := time.Now()
	created, err := queries.CreateUser(ctx, db.CreateUserParams{
		Name:      user.Name().String(),
		Email:     user.Email().String(),
		Password:  user.GetHashedPassword(),
		Status:    db.Status(user.Status()),
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		r.log.Error(ctx, "create user failed", logger.F("error", err))
		return errors.WrapInternalError(err, "create user failed")
	}

	// Update domain object
	user.SetID(int(created.ID))
	user.SetCreatedAt(created.CreatedAt)
	user.SetUpdatedAt(created.UpdatedAt)

	r.log.Info(ctx, "user created successfully", logger.F("user_id", user.ID()))
	return nil
}

// GetByID retrieves a user by ID
func (r *userRepositorySQLC) GetByID(ctx context.Context, id int) (*domain.User, error) {
	r.log.Debug(ctx, "getting user by id", logger.F("user_id", id))

	queries := r.getQueries(ctx)

	user, err := queries.GetUserByID(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, errors.WrapInternalError(err, "get user by id failed")
	}

	return r.dbToDomain(user)
}

// GetByEmail retrieves a user by email
func (r *userRepositorySQLC) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.log.Debug(ctx, "getting user by email", logger.F("email", email))

	queries := r.getQueries(ctx)

	user, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, errors.WrapInternalError(err, "get user by email failed")
	}

	return r.dbToDomain(user)
}

// Update updates a user
func (r *userRepositorySQLC) Update(ctx context.Context, user *domain.User) error {
	r.log.Debug(ctx, "updating user", logger.F("user_id", user.ID()))

	queries := r.getQueries(ctx)

	// Check if user exists
	_, err := r.GetByID(ctx, user.ID())
	if err != nil {
		return err
	}

	// Update user
	updated, err := queries.UpdateUser(ctx, db.UpdateUserParams{
		Name:      user.Name().String(),
		Email:     user.Email().String(),
		Status:    db.Status(user.Status()),
		UpdatedAt: time.Now(),
		ID:        int32(user.ID()),
	})
	if err != nil {
		r.log.Error(ctx, "update user failed", logger.F("error", err))
		return errors.WrapInternalError(err, "update user failed")
	}

	user.SetUpdatedAt(updated.UpdatedAt)

	r.log.Info(ctx, "user updated successfully", logger.F("user_id", user.ID()))
	return nil
}

// Delete deletes a user
func (r *userRepositorySQLC) Delete(ctx context.Context, id int) error {
	r.log.Debug(ctx, "deleting user", logger.F("user_id", id))

	queries := r.getQueries(ctx)

	// Check if user exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete user
	if err := queries.DeleteUser(ctx, int32(id)); err != nil {
		r.log.Error(ctx, "delete user failed", logger.F("error", err))
		return errors.WrapInternalError(err, "delete user failed")
	}

	r.log.Info(ctx, "user deleted successfully", logger.F("user_id", id))
	return nil
}

// List retrieves a paginated list of users
func (r *userRepositorySQLC) List(ctx context.Context, params domain.UserListParams) ([]*domain.User, int64, error) {
	r.log.Debug(ctx, "listing users", logger.F("params", params))

	queries := r.getQueries(ctx)

	// Build query params
	dbParams := db.ListUsersParams{
		ID:     db.ToNullInt32(params.ID),
		Name:   db.ToNullString(params.Name),
		Email:  db.ToNullString(params.Email),
		Status: db.ToNullStatus((*db.Status)(params.Status)),
		Limit:  int32(params.PageSize),
		Offset: int32((params.Page - 1) * params.PageSize),
	}

	// Get total count
	countParams := db.CountUsersParams{
		ID:     dbParams.ID,
		Name:   dbParams.Name,
		Email:  dbParams.Email,
		Status: dbParams.Status,
	}
	total, err := queries.CountUsers(ctx, countParams)
	if err != nil {
		r.log.Error(ctx, "count users failed", logger.F("error", err))
		return nil, 0, errors.WrapInternalError(err, "count users failed")
	}

	// Get users
	users, err := queries.ListUsers(ctx, dbParams)
	if err != nil {
		r.log.Error(ctx, "list users failed", logger.F("error", err))
		return nil, 0, errors.WrapInternalError(err, "list users failed")
	}

	// Convert to domain objects
	result := make([]*domain.User, len(users))
	for i, user := range users {
		domainUser, err := r.dbToDomain(user)
		if err != nil {
			return nil, 0, errors.WrapInternalError(err, "convert domain user failed")
		}
		result[i] = domainUser
	}

	r.log.Debug(ctx, "users listed successfully",
		logger.F("total", total),
		logger.F("count", len(result)))
	return result, total, nil
}

// ExistsByEmail checks if an email exists
func (r *userRepositorySQLC) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	queries := r.getQueries(ctx)
	return queries.ExistsByEmail(ctx, email)
}

// Save saves an aggregate
func (r *userRepositorySQLC) Save(ctx context.Context, aggregate *domain.UserAggregate) error {
	user := aggregate.User()

	if user.ID() == 0 {
		return r.Create(ctx, user)
	}
	return r.Update(ctx, user)
}

// GetAggregateByID retrieves an aggregate by ID
func (r *userRepositorySQLC) GetAggregateByID(ctx context.Context, id int) (*domain.UserAggregate, error) {
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return domain.RebuildUserAggregate(user), nil
}

// GetAggregateByEmail retrieves an aggregate by email
func (r *userRepositorySQLC) GetAggregateByEmail(ctx context.Context, email string) (*domain.UserAggregate, error) {
	user, err := r.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return domain.RebuildUserAggregate(user), nil
}

// dbToDomain converts db.User to domain.User
func (r *userRepositorySQLC) dbToDomain(user db.User) (*domain.User, error) {
	status := domain.Status(user.Status)
	if !status.IsValid() {
		status = domain.StatusInactive
	}

	name, err := domain.NewName(user.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid user name from database: %w", err)
	}

	email, err := domain.NewEmail(user.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid user email from database: %w", err)
	}

	hashedPassword, err := domain.NewHashedPassword(user.Password)
	if err != nil {
		return nil, fmt.Errorf("invalid user password from database: %w", err)
	}

	return domain.NewUser(
		int(user.ID),
		*name,
		*email,
		*hashedPassword,
		status,
		user.CreatedAt,
		user.UpdatedAt,
	)
}

// Ensure implementation
var _ domain.UserRepository = (*userRepositorySQLC)(nil)
