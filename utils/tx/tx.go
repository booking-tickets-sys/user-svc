package tx

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// TxWrapper wraps a database transaction and provides helper methods
type TxWrapper struct {
	tx *sqlx.Tx
}

// NewTxWrapper creates a new transaction wrapper
func NewTxWrapper(tx *sqlx.Tx) *TxWrapper {
	return &TxWrapper{tx: tx}
}

// GetTx returns the underlying transaction
func (tw *TxWrapper) GetTx() *sqlx.Tx {
	return tw.tx
}

// Commit commits the transaction
func (tw *TxWrapper) Commit() error {
	return tw.tx.Commit()
}

// Rollback rolls back the transaction
func (tw *TxWrapper) Rollback() error {
	return tw.tx.Rollback()
}

// ExecContext executes a query with context
func (tw *TxWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return tw.tx.ExecContext(ctx, query, args...)
}

// NamedExecContext executes a named query with context
func (tw *TxWrapper) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	return tw.tx.NamedExecContext(ctx, query, arg)
}

// GetContext gets a single row with context
func (tw *TxWrapper) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return tw.tx.GetContext(ctx, dest, query, args...)
}

// SelectContext selects multiple rows with context
func (tw *TxWrapper) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return tw.tx.SelectContext(ctx, dest, query, args...)
}

// TransactionManager manages database transactions
type TransactionManager struct {
	db *sqlx.DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *sqlx.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// WithTransaction executes a function within a database transaction
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(*TxWrapper) error) error {
	opts := &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	}
	tx, err := tm.db.BeginTxx(ctx, opts)
	if err != nil {
		return err
	}

	txWrapper := NewTxWrapper(tx)

	// Ensure rollback on panic
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after rollback
		}
	}()

	// Execute the function
	if err := fn(txWrapper); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			return err // return original error, not rollback error
		}
		return err
	}

	// Commit on success
	return tx.Commit()
}

// WithTransactionResult executes a function within a database transaction and returns a result
func (tm *TransactionManager) WithTransactionResult(ctx context.Context, fn func(*TxWrapper) (any, error)) (any, error) {
	var result any

	err := tm.WithTransaction(ctx, func(tx *TxWrapper) error {
		var fnErr error
		result, fnErr = fn(tx)
		return fnErr
	})

	return result, err
}
