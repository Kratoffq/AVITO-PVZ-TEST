package repository

import (
	"context"
	"time"

	"github.com/avito/pvz/internal/models"
	"github.com/google/uuid"
)

// Repository определяет интерфейс для работы с хранилищем данных
type Repository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreatePVZ(ctx context.Context, pvz *models.PVZ) error
	GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error)
	GetPVZsWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*PVZWithReceptions, error)
	CreateReception(ctx context.Context, reception *models.Reception) error
	GetReceptionByID(ctx context.Context, id uuid.UUID) (*models.Reception, error)
	GetLastOpenReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	CloseReception(ctx context.Context, receptionID uuid.UUID) error
	CreateProduct(ctx context.Context, product *models.Product) error
	GetLastProduct(ctx context.Context, receptionID uuid.UUID) (*models.Product, error)
	DeleteProduct(ctx context.Context, productID uuid.UUID) error
	CreateProducts(ctx context.Context, products []*models.Product) error
}

// PVZWithReceptions представляет ПВЗ с его приемками
type PVZWithReceptions struct {
	PVZ        *models.PVZ
	Receptions []*ReceptionWithProducts
}

// ReceptionWithProducts представляет приемку с её товарами
type ReceptionWithProducts struct {
	Reception *models.Reception
	Products  []*models.Product
}
