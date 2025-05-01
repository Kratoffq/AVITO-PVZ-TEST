package repository

import (
	"context"

	"github.com/avito/pvz/internal/pvz"
)

// Repository определяет интерфейс для работы с хранилищем PVZ
type Repository interface {
	// Create создает новый PVZ
	Create(ctx context.Context, pvz *pvz.PVZ) error

	// Update обновляет существующий PVZ
	Update(ctx context.Context, pvz *pvz.PVZ) error

	// Delete удаляет PVZ по ID
	Delete(ctx context.Context, id int64) error

	// GetByID возвращает PVZ по ID
	GetByID(ctx context.Context, id int64) (*pvz.PVZ, error)

	// List возвращает список всех PVZ с поддержкой пагинации
	List(ctx context.Context, offset, limit int) ([]*pvz.PVZ, error)

	// GetWithReceptions возвращает PVZ вместе со связанными приемами товаров
	GetWithReceptions(ctx context.Context, id int64) (*pvz.PVZ, []*pvz.Reception, error)

	// GetLast возвращает последний добавленный PVZ
	GetLast(ctx context.Context) (*pvz.PVZ, error)
}

// TransactionManager определяет интерфейс для работы с транзакциями
type TransactionManager interface {
	// BeginTx начинает новую транзакцию
	BeginTx(ctx context.Context) (context.Context, error)

	// CommitTx фиксирует транзакцию
	CommitTx(ctx context.Context) error

	// RollbackTx откатывает транзакцию
	RollbackTx(ctx context.Context) error
}
