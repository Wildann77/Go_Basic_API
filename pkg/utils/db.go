package utils

import (
	"context"

	"gorm.io/gorm"
)

type contextKey string

const (
	// TxKey is the key used to store the transaction in the context
	TxKey contextKey = "tx_key"
)

// GetDBFromContext returns the transaction from the context if it exists,
// otherwise it returns the default db passed as argument.
func GetDBFromContext(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	tx, ok := ctx.Value(TxKey).(*gorm.DB)
	if ok && tx != nil {
		return tx
	}
	return defaultDB.WithContext(ctx)
}

// TransactionFunc is a function that runs within a transaction
type TransactionFunc func(ctx context.Context) error

// RunInTransaction runs the given function within a database transaction.
// It handles commit and rollback automatically.
func RunInTransaction(ctx context.Context, db *gorm.DB, fn TransactionFunc) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Pass the transaction to the context
		txCtx := context.WithValue(ctx, TxKey, tx)
		return fn(txCtx)
	})
}
