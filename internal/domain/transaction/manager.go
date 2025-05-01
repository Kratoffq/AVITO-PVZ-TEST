package transaction

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// SQLManager реализует интерфейс для управления транзакциями
type SQLManager struct {
	db *sqlx.DB
}

// NewManager создает новый экземпляр менеджера транзакций
func NewManager(db *sqlx.DB) Manager {
	return &SQLManager{db: db}
}

// WithinTransaction выполняет функцию в рамках транзакции
func (m *SQLManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Создаем новый контекст с транзакцией
	txCtx := context.WithValue(ctx, TransactionCtxKey{}, tx)

	// Выполняем функцию в рамках транзакции
	if err := fn(txCtx); err != nil {
		return err
	}

	// Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
