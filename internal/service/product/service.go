package product

import (
	"context"
	"errors"
	"time"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/avito/pvz/internal/domain/transaction"
	"github.com/google/uuid"
)

var (
	ErrReceptionNotFound     = errors.New("reception not found")
	ErrReceptionAlreadyClose = errors.New("reception already close")
	ErrProductNotFound       = errors.New("product not found")
	ErrInvalidProductType    = errors.New("invalid product type")
)

// Service определяет бизнес-логику для работы с товарами
type Service struct {
	productRepo   product.Repository
	receptionRepo reception.Repository
	txManager     transaction.Manager
}

// New создает новый экземпляр Service
func New(productRepo product.Repository, receptionRepo reception.Repository, txManager transaction.Manager) *Service {
	return &Service{
		productRepo:   productRepo,
		receptionRepo: receptionRepo,
		txManager:     txManager,
	}
}

// Create создает новый товар
func (s *Service) Create(ctx context.Context, receptionID uuid.UUID, productType product.Type) (*product.Product, error) {
	var result *product.Product

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Проверяем существование приемки
		r, err := s.receptionRepo.GetByID(ctx, receptionID)
		if err != nil {
			return ErrReceptionNotFound
		}

		// Проверяем статус приемки
		if r.Status == reception.StatusClose {
			return ErrReceptionAlreadyClose
		}

		// Валидация типа товара
		if err := validateProductType(productType); err != nil {
			return err
		}

		newProduct := product.New(receptionID, productType)
		if err := s.productRepo.Create(ctx, newProduct); err != nil {
			return err
		}

		result = newProduct
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// CreateBatch создает несколько товаров
func (s *Service) CreateBatch(ctx context.Context, receptionID uuid.UUID, productTypes []product.Type) error {
	return s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Проверяем существование приемки
		r, err := s.receptionRepo.GetByID(ctx, receptionID)
		if err != nil {
			return ErrReceptionNotFound
		}

		// Проверяем статус приемки
		if r.Status == reception.StatusClose {
			return ErrReceptionAlreadyClose
		}

		// Валидация типов товаров
		for _, t := range productTypes {
			if err := validateProductType(t); err != nil {
				return err
			}
		}

		// Создаем товары
		products := make([]*product.Product, len(productTypes))
		for i, t := range productTypes {
			products[i] = product.New(receptionID, t)
		}

		return s.productRepo.CreateBatch(ctx, products)
	})
}

// DeleteLast удаляет последний добавленный товар
func (s *Service) DeleteLast(ctx context.Context, receptionID uuid.UUID) error {
	return s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Проверяем существование приемки
		r, err := s.receptionRepo.GetByID(ctx, receptionID)
		if err != nil {
			return ErrReceptionNotFound
		}

		// Проверяем статус приемки
		if r.Status == reception.StatusClose {
			return ErrReceptionAlreadyClose
		}

		return s.productRepo.DeleteLast(ctx, receptionID)
	})
}

// GetByID получает товар по ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	p, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrProductNotFound
	}
	return p, nil
}

// GetByReceptionID получает все товары приемки
func (s *Service) GetByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]*product.Product, error) {
	return s.productRepo.GetByReceptionID(ctx, receptionID)
}

// List возвращает список товаров
func (s *Service) List(ctx context.Context, offset, limit int) ([]*product.Product, error) {
	return s.productRepo.List(ctx, offset, limit)
}

// validateProductType проверяет корректность типа товара
func validateProductType(t product.Type) error {
	switch t {
	case product.TypeElectronics, product.TypeClothing, product.TypeFood, product.TypeOther:
		return nil
	default:
		return ErrInvalidProductType
	}
}

// AddProduct добавляет новый товар
func (s *Service) AddProduct(ctx context.Context, receptionID uuid.UUID, productType product.Type) (*product.Product, error) {
	product := &product.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        productType,
		ReceptionID: receptionID,
	}

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		return s.productRepo.Create(ctx, product)
	})

	if err != nil {
		return nil, err
	}

	return product, nil
}

// AddProducts добавляет несколько товаров
func (s *Service) AddProducts(ctx context.Context, receptionID uuid.UUID, types []product.Type) ([]*product.Product, error) {
	products := make([]*product.Product, len(types))
	for i, t := range types {
		products[i] = &product.Product{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        t,
			ReceptionID: receptionID,
		}
	}

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		return s.productRepo.CreateBatch(ctx, products)
	})

	if err != nil {
		return nil, err
	}

	return products, nil
}

// DeleteLastProduct удаляет последний добавленный товар
func (s *Service) DeleteLastProduct(ctx context.Context, receptionID uuid.UUID) error {
	return s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		return s.productRepo.DeleteLast(ctx, receptionID)
	})
}

// GetProducts получает все товары приемки
func (s *Service) GetProducts(ctx context.Context, receptionID uuid.UUID) ([]*product.Product, error) {
	return s.productRepo.GetByReceptionID(ctx, receptionID)
}
