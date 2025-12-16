package database

import (
	"context"

	"gorm.io/gorm"
)

type contextKey string

const txKey contextKey = "db_tx"

//go:generate mockgen -source=$GOFILE -destination=../../internal/mocks/tx_manager_mock.go -package=mocks
// TransactionManager handles database transactions.
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// gormTransactionManager implements TransactionManager for GORM.
type gormTransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager creates a new GORM transaction manager.
func NewTransactionManager(db *gorm.DB) TransactionManager {
	return &gormTransactionManager{db: db}
}

// WithTransaction runs the given function within a database transaction.
func (tm *gormTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Store the transaction object in the context
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}

// GetDBFromContext retrieves the transaction DB from context if present, otherwise returns the default DB.
func GetDBFromContext(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	tx, ok := ctx.Value(txKey).(*gorm.DB)
	if ok {
		return tx
	}
	return defaultDB.WithContext(ctx)
}
