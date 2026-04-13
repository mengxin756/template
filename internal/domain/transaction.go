package domain

import (
	"context"
)

// TransactionManager manages database transactions
// This interface belongs to domain layer, implemented by infrastructure layer
type TransactionManager interface {
	// WithTransaction executes a function within a transaction
	// If the function returns an error, the transaction is rolled back
	// If the function returns nil, the transaction is committed
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error

	// WithTransactionResult executes a function within a transaction and returns a result
	WithTransactionResult(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)
}

// TransactionalRepository indicates a repository that supports transactions
// Repositories can accept a transactional context to participate in ongoing transactions
type TransactionalRepository interface {
	// WithTx returns a new repository instance that uses the provided transaction context
	WithTx(ctx context.Context) TransactionalRepository
}

// TxContextKey is the context key for transaction context
type TxContextKey struct{}

// ContextWithTx returns a new context with the transaction client
func ContextWithTx(ctx context.Context, tx interface{}) context.Context {
	return context.WithValue(ctx, TxContextKey{}, tx)
}

// TxFromContext retrieves the transaction client from context
func TxFromContext(ctx context.Context) interface{} {
	if tx, ok := ctx.Value(TxContextKey{}).(interface{}); ok {
		return tx
	}
	return nil
}
