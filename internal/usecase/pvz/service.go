package pvz

import (
	"context"

	"github.com/avito/pvz/internal/domain/pvz"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/avito/pvz/internal/domain/transaction"
	"github.com/google/uuid"
)

// UseCase определяет методы бизнес-логики для работы с ПВЗ
type UseCase interface {
	// CreatePVZ создает новый ПВЗ
	CreatePVZ(ctx context.Context, req *CreatePVZRequest) (*CreatePVZResponse, error)

	// GetPVZWithReceptions получает ПВЗ с приемками за период
	GetPVZWithReceptions(ctx context.Context, req *GetPVZWithReceptionsRequest) (*GetPVZWithReceptionsResponse, error)

	// CreateReception создает новую приемку для ПВЗ
	CreateReception(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error)
}

type useCase struct {
	pvzRepo       pvz.Repository
	receptionRepo reception.Repository
	txManager     transaction.Manager
}

// New создает новый экземпляр UseCase
func New(
	pvzRepo pvz.Repository,
	receptionRepo reception.Repository,
	txManager transaction.Manager,
) UseCase {
	return &useCase{
		pvzRepo:       pvzRepo,
		receptionRepo: receptionRepo,
		txManager:     txManager,
	}
}

func (u *useCase) CreatePVZ(ctx context.Context, req *CreatePVZRequest) (*CreatePVZResponse, error) {
	var result *pvz.PVZ

	err := u.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		newPVZ := pvz.New(req.City)
		if err := u.pvzRepo.Create(ctx, newPVZ); err != nil {
			return err
		}
		result = newPVZ
		return nil
	})

	if err != nil {
		return nil, err
	}

	return ToPVZResponse(result), nil
}

func (u *useCase) GetPVZWithReceptions(
	ctx context.Context,
	req *GetPVZWithReceptionsRequest,
) (*GetPVZWithReceptionsResponse, error) {
	items, err := u.pvzRepo.GetWithReceptions(ctx, req.StartDate, req.EndDate, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	return ToPVZWithReceptionsResponse(items), nil
}

func (u *useCase) CreateReception(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	var result *reception.Reception

	err := u.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Проверяем существование ПВЗ
		if _, err := u.pvzRepo.GetByID(ctx, pvzID); err != nil {
			return err
		}

		// Проверяем, нет ли уже открытой приемки
		if _, err := u.receptionRepo.GetOpenByPVZID(ctx, pvzID); err != reception.ErrNoOpenReception {
			if err == nil {
				return reception.ErrReceptionAlreadyOpen
			}
			return err
		}

		newReception := reception.New(pvzID)
		if err := u.receptionRepo.Create(ctx, newReception); err != nil {
			return err
		}

		result = newReception
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
