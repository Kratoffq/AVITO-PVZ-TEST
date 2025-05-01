package reception

import (
	"context"
	"errors"
	"time"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/avito/pvz/internal/domain/pvz"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/avito/pvz/internal/domain/transaction"
	"github.com/avito/pvz/internal/metrics"
	"github.com/google/uuid"
)

var (
	ErrPVZNotFound           = errors.New("pvz not found")
	ErrReceptionNotFound     = errors.New("reception not found")
	ErrReceptionAlreadyOpen  = errors.New("reception already open")
	ErrReceptionAlreadyClose = errors.New("reception already close")
)

// Service определяет бизнес-логику для работы с приемками
type Service struct {
	receptionRepo reception.Repository
	pvzRepo       pvz.Repository
	txManager     transaction.Manager
	productRepo   product.Repository
}

// New создает новый экземпляр Service
func New(receptionRepo reception.Repository, pvzRepo pvz.Repository, txManager transaction.Manager, productRepo product.Repository) *Service {
	return &Service{
		receptionRepo: receptionRepo,
		pvzRepo:       pvzRepo,
		txManager:     txManager,
		productRepo:   productRepo,
	}
}

// Create создает новую приемку
func (s *Service) Create(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	start := time.Now()
	var result *reception.Reception

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Проверяем существование ПВЗ
		if _, err := s.pvzRepo.GetByID(ctx, pvzID); err != nil {
			return ErrPVZNotFound
		}

		// Проверяем, нет ли уже открытой приемки
		if _, err := s.receptionRepo.GetOpenByPVZID(ctx, pvzID); err != reception.ErrNoOpenReception {
			if err == nil {
				return ErrReceptionAlreadyOpen
			}
			return err
		}

		newReception := reception.New(pvzID)
		if err := s.receptionRepo.Create(ctx, newReception); err != nil {
			return err
		}

		result = newReception
		return nil
	})

	// Обновляем метрики
	metrics.TransactionDuration.WithLabelValues("create_reception").Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.TransactionErrors.WithLabelValues("create_reception").Inc()
		return nil, err
	}

	metrics.ReceptionCreatedTotal.Inc()
	return result, nil
}

// Close закрывает приемку
func (s *Service) Close(ctx context.Context, pvzID uuid.UUID) error {
	return s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		reception, err := s.receptionRepo.GetLastOpen(ctx, pvzID)
		if err != nil {
			return err
		}

		reception.Status = "close"
		return s.receptionRepo.Update(ctx, reception)
	})
}

// GetByID получает приемку по ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*reception.Reception, error) {
	start := time.Now()
	r, err := s.receptionRepo.GetByID(ctx, id)

	// Обновляем метрики
	metrics.TransactionDuration.WithLabelValues("get_reception").Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.TransactionErrors.WithLabelValues("get_reception").Inc()
	}

	if err != nil {
		return nil, ErrReceptionNotFound
	}
	return r, nil
}

// GetOpenByPVZID получает открытую приемку для ПВЗ
func (s *Service) GetOpenByPVZID(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	r, err := s.receptionRepo.GetOpenByPVZID(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// List возвращает список приемок
func (s *Service) List(ctx context.Context, offset, limit int) ([]*reception.Reception, error) {
	return s.receptionRepo.List(ctx, offset, limit)
}

// GetProducts получает список товаров приемки
func (s *Service) GetProducts(ctx context.Context, receptionID uuid.UUID) ([]*product.Product, error) {
	return s.receptionRepo.GetProducts(ctx, receptionID)
}

// CreateProduct добавляет товар в приемку
func (s *Service) CreateProduct(ctx context.Context, receptionID uuid.UUID, productType string) error {
	start := time.Now()

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Получаем приемку
		r, err := s.receptionRepo.GetByID(ctx, receptionID)
		if err != nil {
			return ErrReceptionNotFound
		}

		// Проверяем статус
		if r.Status == reception.StatusClose {
			return ErrReceptionAlreadyClose
		}

		// Создаем товар
		p := product.New(r.ID, product.Type(productType))
		return s.productRepo.Create(ctx, p)
	})

	// Обновляем метрики
	metrics.TransactionDuration.WithLabelValues("create_product").Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.TransactionErrors.WithLabelValues("create_product").Inc()
		return err
	}

	metrics.ProductCreatedTotal.Inc()
	return nil
}
