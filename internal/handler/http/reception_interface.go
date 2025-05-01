package http

import (
	"context"

	"github.com/avito/pvz/internal/domain/reception"
	"github.com/google/uuid"
)

// ReceptionServiceInterface определяет интерфейс для сервиса приемок
type ReceptionServiceInterface interface {
	Create(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error)
	GetByID(ctx context.Context, id uuid.UUID) (*reception.Reception, error)
	Close(ctx context.Context, pvzID uuid.UUID) error
	GetOpenByPVZID(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error)
	List(ctx context.Context, offset, limit int) ([]*reception.Reception, error)
}
