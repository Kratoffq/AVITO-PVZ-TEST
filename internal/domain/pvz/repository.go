package pvz

import (
	"context"
	"time"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/google/uuid"
)

// Repository определяет методы для работы с ПВЗ в хранилище
type Repository interface {
	// Create создает новый ПВЗ
	Create(ctx context.Context, pvz *PVZ) error

	// GetByID получает ПВЗ по ID
	GetByID(ctx context.Context, id uuid.UUID) (*PVZ, error)

	// GetByCity получает ПВЗ по городу
	GetByCity(ctx context.Context, city string) (*PVZ, error)

	// Update обновляет данные ПВЗ
	Update(ctx context.Context, pvz *PVZ) error

	// Delete удаляет ПВЗ по ID
	Delete(ctx context.Context, id uuid.UUID) error

	// List возвращает список ПВЗ с пагинацией
	List(ctx context.Context, offset, limit int) ([]*PVZ, error)

	// GetWithReceptions получает список ПВЗ с приемками за период
	GetWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*PVZWithReceptions, error)

	// GetAll возвращает список всех ПВЗ
	GetAll(ctx context.Context) ([]*PVZ, error)
}

// PVZWithReceptions представляет ПВЗ с его приемками
type PVZWithReceptions struct {
	PVZ        *PVZ
	Receptions []*ReceptionWithProducts
}

// ReceptionWithProducts представляет приемку с товарами
type ReceptionWithProducts struct {
	Reception *reception.Reception
	Products  []*product.Product
}
