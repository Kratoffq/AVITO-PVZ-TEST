package reception

import (
	"context"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/google/uuid"
)

// Repository определяет методы для работы с приемками в хранилище
type Repository interface {
	// Create создает новую приемку
	Create(ctx context.Context, reception *Reception) error

	// GetByID получает приемку по ID
	GetByID(ctx context.Context, id uuid.UUID) (*Reception, error)

	// Update обновляет данные приемки
	Update(ctx context.Context, reception *Reception) error

	// Delete удаляет приемку по ID
	Delete(ctx context.Context, id uuid.UUID) error

	// List возвращает список приемок с пагинацией
	List(ctx context.Context, offset, limit int) ([]*Reception, error)

	// GetOpenByPVZID получает открытую приемку для ПВЗ
	GetOpenByPVZID(ctx context.Context, pvzID uuid.UUID) (*Reception, error)

	// GetProducts получает список товаров приемки
	GetProducts(ctx context.Context, receptionID uuid.UUID) ([]*product.Product, error)

	// GetLastOpen получает последнюю открытую приемку для ПВЗ
	GetLastOpen(ctx context.Context, pvzID uuid.UUID) (*Reception, error)
}
