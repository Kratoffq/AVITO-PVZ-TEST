package transaction

import "context"

// Manager определяет интерфейс для управления транзакциями
type Manager interface {
	// WithinTransaction выполняет функцию в рамках транзакции
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// TransactionCtxKey - ключ для хранения транзакции в контексте
type TransactionCtxKey struct{}
