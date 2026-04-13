package membership

import (
	"context"
)

// MemberRepository member repository interface
type MemberRepository interface {
	// Create creates member
	Create(ctx context.Context, member *Member) error

	// GetByID gets member by ID
	GetByID(ctx context.Context, id int) (*Member, error)

	// GetByUserID gets member by user ID
	GetByUserID(ctx context.Context, userID int) (*Member, error)

	// Update updates member
	Update(ctx context.Context, member *Member) error

	// Save saves member aggregate
	Save(ctx context.Context, aggregate *MemberAggregate) error

	// GetAggregateByID gets member aggregate by ID
	GetAggregateByID(ctx context.Context, id int) (*MemberAggregate, error)

	// GetAggregateByUserID gets member aggregate by user ID
	GetAggregateByUserID(ctx context.Context, userID int) (*MemberAggregate, error)

	// List lists members by query
	List(ctx context.Context, query *MemberQuery) ([]*Member, int64, error)
}

// MemberQuery member query conditions
type MemberQuery struct {
	ID      *int
	UserID  *int
	Level   *MemberLevel
	Status  *MemberStatus
	Page    int
	PageSize int
}
