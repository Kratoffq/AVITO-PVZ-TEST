package service

import (
	"context"
	"time"

	"github.com/avito/pvz/internal/models"
	"github.com/avito/pvz/internal/repository"
	"github.com/google/uuid"
)

type Service interface {
	// User methods
	RegisterUser(ctx context.Context, email, password string, role models.UserRole) (*models.User, error)
	LoginUser(ctx context.Context, email, password string) (string, error)
	DummyLogin(ctx context.Context, role models.UserRole) (string, error)

	// PVZ methods
	CreatePVZ(ctx context.Context, city string) (*models.PVZ, error)
	GetPVZsWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*repository.PVZWithReceptions, error)

	// Reception methods
	CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	CloseReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)

	// Product methods
	AddProduct(ctx context.Context, pvzID uuid.UUID, productType models.ProductType) (*models.Product, error)
	AddProducts(ctx context.Context, pvzID uuid.UUID, productTypes []models.ProductType) ([]*models.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
}
