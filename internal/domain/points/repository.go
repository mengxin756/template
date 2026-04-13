package points

import (
	"context"
)

// PointsAccountRepository points account repository interface
type PointsAccountRepository interface {
	// Create creates points account
	Create(ctx context.Context, account *PointsAccount) error

	// GetByID gets points account by ID
	GetByID(ctx context.Context, id int) (*PointsAccount, error)

	// GetByMemberID gets points account by member ID
	GetByMemberID(ctx context.Context, memberID int) (*PointsAccount, error)

	// GetByUserID gets points account by user ID
	GetByUserID(ctx context.Context, userID int) (*PointsAccountAggregate, error)

	// Update updates points account
	Update(ctx context.Context, account *PointsAccount) error

	// Save saves points account aggregate
	Save(ctx context.Context, aggregate *PointsAccountAggregate) error

	// GetAggregateByID gets points account aggregate by ID
	GetAggregateByID(ctx context.Context, id int) (*PointsAccountAggregate, error)

	// GetAggregateByMemberID gets points account aggregate by member ID
	GetAggregateByMemberID(ctx context.Context, memberID int) (*PointsAccountAggregate, error)

	// CreateRecord creates points record
	CreateRecord(ctx context.Context, record *PointsRecord) error

	// GetRecordsByAccountID gets points records by account ID
	GetRecordsByAccountID(ctx context.Context, accountID int, limit, offset int) ([]*PointsRecord, error)
}

// PointsAccountQuery points account query conditions
type PointsAccountQuery struct {
	ID       *int
	MemberID *int
	Page     int
	PageSize int
}
