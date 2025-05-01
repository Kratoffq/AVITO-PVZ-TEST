package reception

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/avito/pvz/internal/domain/pvz"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockReceptionRepository реализует мок для reception.Repository
type MockReceptionRepository struct {
	mock.Mock
}

func (m *MockReceptionRepository) Create(ctx context.Context, r *reception.Reception) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockReceptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*reception.Reception, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception.Reception), args.Error(1)
}

func (m *MockReceptionRepository) Update(ctx context.Context, r *reception.Reception) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockReceptionRepository) List(ctx context.Context, offset, limit int) ([]*reception.Reception, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*reception.Reception), args.Error(1)
}

func (m *MockReceptionRepository) GetOpenByPVZID(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception.Reception), args.Error(1)
}

func (m *MockReceptionRepository) GetProducts(ctx context.Context, receptionID uuid.UUID) ([]*product.Product, error) {
	args := m.Called(ctx, receptionID)
	return args.Get(0).([]*product.Product), args.Error(1)
}

func (m *MockReceptionRepository) GetLastOpen(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception.Reception), args.Error(1)
}

func (m *MockReceptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockPVZRepository реализует мок для pvz.Repository
type MockPVZRepository struct {
	mock.Mock
}

func (m *MockPVZRepository) Create(ctx context.Context, p *pvz.PVZ) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPVZRepository) GetByID(ctx context.Context, id uuid.UUID) (*pvz.PVZ, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) Update(ctx context.Context, p *pvz.PVZ) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPVZRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPVZRepository) List(ctx context.Context, offset, limit int) ([]*pvz.PVZ, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) GetAll(ctx context.Context) ([]*pvz.PVZ, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) GetWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*pvz.PVZWithReceptions, error) {
	args := m.Called(ctx, startDate, endDate, page, limit)
	return args.Get(0).([]*pvz.PVZWithReceptions), args.Error(1)
}

func (m *MockPVZRepository) GetByCity(ctx context.Context, city string) (*pvz.PVZ, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pvz.PVZ), args.Error(1)
}

// MockProductRepository реализует мок для product.Repository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, p *product.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, p *product.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) List(ctx context.Context, offset, limit int) ([]*product.Product, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*product.Product), args.Error(1)
}

func (m *MockProductRepository) CreateBatch(ctx context.Context, products []*product.Product) error {
	args := m.Called(ctx, products)
	return args.Error(0)
}

func (m *MockProductRepository) DeleteLast(ctx context.Context, receptionID uuid.UUID) error {
	args := m.Called(ctx, receptionID)
	return args.Error(0)
}

func (m *MockProductRepository) GetByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]*product.Product, error) {
	args := m.Called(ctx, receptionID)
	return args.Get(0).([]*product.Product), args.Error(1)
}

func (m *MockProductRepository) GetLast(ctx context.Context, receptionID uuid.UUID) (*product.Product, error) {
	args := m.Called(ctx, receptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

// MockTransactionManager реализует мок для transaction.Manager
type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	if fn != nil {
		fn(ctx)
	}
	return args.Error(0)
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name          string
		pvzID         uuid.UUID
		setupMocks    func(*MockReceptionRepository, *MockPVZRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name:  "успешное создание",
			pvzID: uuid.New(),
			setupMocks: func(receptionRepo *MockReceptionRepository, pvzRepo *MockPVZRepository, tx *MockTransactionManager) {
				pvzRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&pvz.PVZ{}, nil)
				receptionRepo.On("GetOpenByPVZID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, reception.ErrNoOpenReception)
				receptionRepo.On("Create", mock.Anything, mock.AnythingOfType("*reception.Reception")).Return(nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "пвз не найден",
			pvzID: uuid.New(),
			setupMocks: func(receptionRepo *MockReceptionRepository, pvzRepo *MockPVZRepository, tx *MockTransactionManager) {
				pvzRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("not found"))
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrPVZNotFound)
			},
			expectedError: ErrPVZNotFound,
		},
		{
			name:  "приемка уже открыта",
			pvzID: uuid.New(),
			setupMocks: func(receptionRepo *MockReceptionRepository, pvzRepo *MockPVZRepository, tx *MockTransactionManager) {
				pvzRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&pvz.PVZ{}, nil)
				receptionRepo.On("GetOpenByPVZID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&reception.Reception{}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrReceptionAlreadyOpen)
			},
			expectedError: ErrReceptionAlreadyOpen,
		},
		{
			name:  "ошибка при создании приемки",
			pvzID: uuid.New(),
			setupMocks: func(receptionRepo *MockReceptionRepository, pvzRepo *MockPVZRepository, tx *MockTransactionManager) {
				pvzRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&pvz.PVZ{}, nil)
				receptionRepo.On("GetOpenByPVZID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, reception.ErrNoOpenReception)
				receptionRepo.On("Create", mock.Anything, mock.AnythingOfType("*reception.Reception")).Return(errors.New("database error"))
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receptionRepo := new(MockReceptionRepository)
			pvzRepo := new(MockPVZRepository)
			productRepo := new(MockProductRepository)
			tx := new(MockTransactionManager)
			tt.setupMocks(receptionRepo, pvzRepo, tx)

			service := New(receptionRepo, pvzRepo, tx, productRepo)
			_, err := service.Create(context.Background(), tt.pvzID)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			receptionRepo.AssertExpectations(t)
			pvzRepo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}

func TestService_Close(t *testing.T) {
	tests := []struct {
		name          string
		pvzID         uuid.UUID
		setupMocks    func(*MockReceptionRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name:  "успешное закрытие",
			pvzID: uuid.New(),
			setupMocks: func(receptionRepo *MockReceptionRepository, tx *MockTransactionManager) {
				receptionRepo.On("GetLastOpen", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&reception.Reception{}, nil)
				receptionRepo.On("Update", mock.Anything, mock.AnythingOfType("*reception.Reception")).Return(nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "приемка не найдена",
			pvzID: uuid.New(),
			setupMocks: func(receptionRepo *MockReceptionRepository, tx *MockTransactionManager) {
				receptionRepo.On("GetLastOpen", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("not found"))
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(errors.New("not found"))
			},
			expectedError: errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receptionRepo := new(MockReceptionRepository)
			tx := new(MockTransactionManager)
			tt.setupMocks(receptionRepo, tx)

			service := New(receptionRepo, nil, tx, nil)
			err := service.Close(context.Background(), tt.pvzID)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			receptionRepo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		id            uuid.UUID
		setupMocks    func(*MockReceptionRepository)
		expectedError error
	}{
		{
			name: "успешное получение",
			id:   uuid.New(),
			setupMocks: func(repo *MockReceptionRepository) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&reception.Reception{}, nil)
			},
			expectedError: nil,
		},
		{
			name: "приемка не найдена",
			id:   uuid.New(),
			setupMocks: func(repo *MockReceptionRepository) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("not found"))
			},
			expectedError: ErrReceptionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receptionRepo := new(MockReceptionRepository)
			tt.setupMocks(receptionRepo)

			service := New(receptionRepo, nil, nil, nil)
			_, err := service.GetByID(context.Background(), tt.id)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			receptionRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetOpenByPVZID(t *testing.T) {
	tests := []struct {
		name          string
		pvzID         uuid.UUID
		setupMocks    func(*MockReceptionRepository)
		expectedError error
	}{
		{
			name:  "успешное получение",
			pvzID: uuid.New(),
			setupMocks: func(repo *MockReceptionRepository) {
				repo.On("GetOpenByPVZID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&reception.Reception{}, nil)
			},
			expectedError: nil,
		},
		{
			name:  "приемка не найдена",
			pvzID: uuid.New(),
			setupMocks: func(repo *MockReceptionRepository) {
				repo.On("GetOpenByPVZID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("not found"))
			},
			expectedError: errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receptionRepo := new(MockReceptionRepository)
			tt.setupMocks(receptionRepo)

			service := New(receptionRepo, nil, nil, nil)
			_, err := service.GetOpenByPVZID(context.Background(), tt.pvzID)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			receptionRepo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	tests := []struct {
		name          string
		offset        int
		limit         int
		setupMocks    func(*MockReceptionRepository)
		expectedError error
	}{
		{
			name:   "успешное получение списка",
			offset: 0,
			limit:  10,
			setupMocks: func(repo *MockReceptionRepository) {
				repo.On("List", mock.Anything, 0, 10).Return([]*reception.Reception{}, nil)
			},
			expectedError: nil,
		},
		{
			name:   "ошибка при получении списка",
			offset: 0,
			limit:  10,
			setupMocks: func(repo *MockReceptionRepository) {
				repo.On("List", mock.Anything, 0, 10).Return([]*reception.Reception(nil), errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receptionRepo := new(MockReceptionRepository)
			tt.setupMocks(receptionRepo)

			service := New(receptionRepo, nil, nil, nil)
			_, err := service.List(context.Background(), tt.offset, tt.limit)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			receptionRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetProducts(t *testing.T) {
	tests := []struct {
		name          string
		receptionID   uuid.UUID
		setupMocks    func(*MockReceptionRepository)
		expectedError error
	}{
		{
			name:        "успешное получение списка товаров",
			receptionID: uuid.New(),
			setupMocks: func(repo *MockReceptionRepository) {
				repo.On("GetProducts", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return([]*product.Product{}, nil)
			},
			expectedError: nil,
		},
		{
			name:        "ошибка при получении списка товаров",
			receptionID: uuid.New(),
			setupMocks: func(repo *MockReceptionRepository) {
				repo.On("GetProducts", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return([]*product.Product(nil), errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receptionRepo := new(MockReceptionRepository)
			tt.setupMocks(receptionRepo)

			service := New(receptionRepo, nil, nil, nil)
			_, err := service.GetProducts(context.Background(), tt.receptionID)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			receptionRepo.AssertExpectations(t)
		})
	}
}

func TestService_CreateProduct(t *testing.T) {
	tests := []struct {
		name          string
		receptionID   uuid.UUID
		productType   string
		setupMocks    func(*MockReceptionRepository, *MockProductRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name:        "успешное создание товара",
			receptionID: uuid.New(),
			productType: string(product.TypeElectronics),
			setupMocks: func(receptionRepo *MockReceptionRepository, productRepo *MockProductRepository, tx *MockTransactionManager) {
				receptionRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&reception.Reception{Status: reception.StatusInProgress}, nil)
				productRepo.On("Create", mock.Anything, mock.AnythingOfType("*product.Product")).Return(nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "приемка не найдена",
			receptionID: uuid.New(),
			productType: string(product.TypeElectronics),
			setupMocks: func(receptionRepo *MockReceptionRepository, productRepo *MockProductRepository, tx *MockTransactionManager) {
				receptionRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("not found"))
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrReceptionNotFound)
			},
			expectedError: ErrReceptionNotFound,
		},
		{
			name:        "приемка закрыта",
			receptionID: uuid.New(),
			productType: string(product.TypeElectronics),
			setupMocks: func(receptionRepo *MockReceptionRepository, productRepo *MockProductRepository, tx *MockTransactionManager) {
				receptionRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&reception.Reception{Status: reception.StatusClose}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrReceptionAlreadyClose)
			},
			expectedError: ErrReceptionAlreadyClose,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receptionRepo := new(MockReceptionRepository)
			productRepo := new(MockProductRepository)
			tx := new(MockTransactionManager)
			tt.setupMocks(receptionRepo, productRepo, tx)

			service := New(receptionRepo, nil, tx, productRepo)
			err := service.CreateProduct(context.Background(), tt.receptionID, tt.productType)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			receptionRepo.AssertExpectations(t)
			productRepo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}
