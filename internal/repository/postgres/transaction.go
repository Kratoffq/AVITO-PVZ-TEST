package postgres

import (
	"context"
	"database/sql"
)

// TransactionManager реализует интерфейс transaction.Manager
type TransactionManager struct {
	db *sql.DB
}

// NewTransactionManager создает новый экземпляр TransactionManager
func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// WithinTransaction выполняет функцию в транзакции
func (tm *TransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
