package wallet

import (
	"context"
)

// WalletRepository wallet repository interface
type WalletRepository interface {
	// Create creates wallet
	Create(ctx context.Context, wallet *Wallet) error

	// GetByID gets wallet by ID
	GetByID(ctx context.Context, id int) error

	// GetByUserID gets wallet by user ID
	GetByUserID(ctx context.Context, userID int) (*Wallet, error)

	// Update updates wallet
	Update(ctx context.Context, wallet *Wallet) error

	// Save saves wallet aggregate
	Save(ctx context.Context, aggregate *WalletAggregate) error

	// GetAggregateByID gets wallet aggregate by ID
	GetAggregateByID(ctx context.Context, id int) (*WalletAggregate, error)

	// GetAggregateByUserID gets wallet aggregate by user ID
	GetAggregateByUserID(ctx context.Context, userID int) (*WalletAggregate, error)

	// CreateTransaction creates transaction
	CreateTransaction(ctx context.Context, transaction *Transaction) error

	// GetTransactionsByWalletID gets transactions by wallet ID
	GetTransactionsByWalletID(ctx context.Context, walletID int, limit, offset int) ([]*Transaction, error)
}

// WalletQuery wallet query conditions
type WalletQuery struct {
	ID      *int
	UserID  *int
	Page    int
	PageSize int
}
