package data

import (
	"context"
	"database/sql"
	"fmt"

	"example.com/classic/internal/domain"
	"example.com/classic/pkg/logger"
)

// TransactionManager implements domain.TransactionManager using standard sql.DB
type TransactionManager struct {
	db  *sql.DB
	log logger.Logger
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *sql.DB, log logger.Logger) *TransactionManager {
	return &TransactionManager{
		db:  db,
		log: log,
	}
}

// WithTransaction executes a function within a transaction
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Create transaction context
	txCtx := domain.ContextWithTx(ctx, tx)

	// Execute function
	if err := fn(txCtx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			tm.log.Error(ctx, "failed to rollback transaction",
				logger.F("error", rollbackErr),
				logger.F("original_error", err))
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// WithTransactionResult executes a function within a transaction and returns a result
func (tm *TransactionManager) WithTransactionResult(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	var result interface{}

	err := tm.WithTransaction(ctx, func(ctx context.Context) error {
		r, err := fn(ctx)
		if err != nil {
			return err
		}
		result = r
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Ensure TransactionManager implements domain.TransactionManager
var _ domain.TransactionManager = (*TransactionManager)(nil)
