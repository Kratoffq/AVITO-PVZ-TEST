package product

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Repository определяет методы для работы с товарами в хранилище
type Repository interface {
	// Create создает новый товар
	Create(ctx context.Context, product *Product) error

	// CreateBatch создает несколько товаров в одной транзакции
	CreateBatch(ctx context.Context, products []*Product) error

	// GetByID получает товар по ID
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)

	// Update обновляет данные товара
	Update(ctx context.Context, product *Product) error

	// Delete удаляет товар по ID
	Delete(ctx context.Context, id uuid.UUID) error

	// List возвращает список товаров с пагинацией
	List(ctx context.Context, offset, limit int) ([]*Product, error)

	// GetLast получает последний добавленный товар из приемки
	GetLast(ctx context.Context, receptionID uuid.UUID) (*Product, error)

	// DeleteLast удаляет последний добавленный товар из приемки
	DeleteLast(ctx context.Context, receptionID uuid.UUID) error

	// GetByReceptionID получает все товары приемки
	GetByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]*Product, error)
}

// ErrProductNotFound возвращается, когда товар не найден
var ErrProductNotFound = errors.New("product not found")
