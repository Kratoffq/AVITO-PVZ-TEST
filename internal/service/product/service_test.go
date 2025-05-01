package product

import (
	"context"
	"errors"
	"testing"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductRepository реализует мок для product.Repository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, p *product.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockProductRepository) CreateBatch(ctx context.Context, products []*product.Product) error {
	args := m.Called(ctx, products)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*product.Product), args.Error(1)
}

func (m *MockProductRepository) GetLast(ctx context.Context, receptionID uuid.UUID) (*product.Product, error) {
	args := m.Called(ctx, receptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

func (m *MockProductRepository) DeleteLast(ctx context.Context, receptionID uuid.UUID) error {
	args := m.Called(ctx, receptionID)
	return args.Error(0)
}

func (m *MockProductRepository) GetByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]*product.Product, error) {
	args := m.Called(ctx, receptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*product.Product), args.Error(1)
}

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

func (m *MockReceptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReceptionRepository) List(ctx context.Context, offset, limit int) ([]*reception.Reception, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*product.Product), args.Error(1)
}

func (m *MockReceptionRepository) GetLastOpen(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception.Reception), args.Error(1)
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

func TestService_CreateBatch(t *testing.T) {
	tests := []struct {
		name          string
		receptionID   uuid.UUID
		productTypes  []product.Type
		setupMocks    func(*MockProductRepository, *MockReceptionRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name:         "успешное создание",
			receptionID:  uuid.New(),
			productTypes: []product.Type{product.TypeElectronics, product.TypeClothing},
			setupMocks: func(productRepo *MockProductRepository, receptionRepo *MockReceptionRepository, tx *MockTransactionManager) {
				receptionRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&reception.Reception{Status: reception.StatusInProgress}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
				productRepo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*product.Product")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:         "приемка не найдена",
			receptionID:  uuid.New(),
			productTypes: []product.Type{product.TypeElectronics},
			setupMocks: func(productRepo *MockProductRepository, receptionRepo *MockReceptionRepository, tx *MockTransactionManager) {
				receptionRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("not found"))
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrReceptionNotFound)
			},
			expectedError: ErrReceptionNotFound,
		},
		{
			name:         "приемка закрыта",
			receptionID:  uuid.New(),
			productTypes: []product.Type{product.TypeElectronics},
			setupMocks: func(productRepo *MockProductRepository, receptionRepo *MockReceptionRepository, tx *MockTransactionManager) {
				receptionRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&reception.Reception{Status: reception.StatusClose}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrReceptionAlreadyClose)
			},
			expectedError: ErrReceptionAlreadyClose,
		},
		{
			name:         "некорректный тип товара",
			receptionID:  uuid.New(),
			productTypes: []product.Type{product.Type("invalid")},
			setupMocks: func(productRepo *MockProductRepository, receptionRepo *MockReceptionRepository, tx *MockTransactionManager) {
				receptionRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&reception.Reception{Status: reception.StatusInProgress}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrInvalidProductType)
			},
			expectedError: ErrInvalidProductType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(MockProductRepository)
			receptionRepo := new(MockReceptionRepository)
			tx := new(MockTransactionManager)
			tt.setupMocks(productRepo, receptionRepo, tx)

			service := New(productRepo, receptionRepo, tx)
			err := service.CreateBatch(context.Background(), tt.receptionID, tt.productTypes)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			productRepo.AssertExpectations(t)
			receptionRepo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}

func TestService_DeleteLast(t *testing.T) {
	tests := []struct {
		name          string
		receptionID   uuid.UUID
		setupMocks    func(*MockProductRepository, *MockReceptionRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name:        "успешное удаление",
			receptionID: uuid.New(),
			setupMocks: func(productRepo *MockProductRepository, receptionRepo *MockReceptionRepository, tx *MockTransactionManager) {
				receptionRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&reception.Reception{Status: reception.StatusInProgress}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
				productRepo.On("DeleteLast", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "приемка не найдена",
			receptionID: uuid.New(),
			setupMocks: func(productRepo *MockProductRepository, receptionRepo *MockReceptionRepository, tx *MockTransactionManager) {
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
			setupMocks: func(productRepo *MockProductRepository, receptionRepo *MockReceptionRepository, tx *MockTransactionManager) {
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
			productRepo := new(MockProductRepository)
			receptionRepo := new(MockReceptionRepository)
			tx := new(MockTransactionManager)
			tt.setupMocks(productRepo, receptionRepo, tx)

			service := New(productRepo, receptionRepo, tx)
			err := service.DeleteLast(context.Background(), tt.receptionID)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			productRepo.AssertExpectations(t)
			receptionRepo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		id            uuid.UUID
		setupMocks    func(*MockProductRepository)
		expectedError error
	}{
		{
			name: "успешное получение",
			id:   uuid.New(),
			setupMocks: func(repo *MockProductRepository) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&product.Product{}, nil)
			},
			expectedError: nil,
		},
		{
			name: "товар не найден",
			id:   uuid.New(),
			setupMocks: func(repo *MockProductRepository) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("not found"))
			},
			expectedError: ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(MockProductRepository)
			tt.setupMocks(productRepo)

			service := New(productRepo, nil, nil)
			_, err := service.GetByID(context.Background(), tt.id)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			productRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByReceptionID(t *testing.T) {
	tests := []struct {
		name          string
		receptionID   uuid.UUID
		setupMocks    func(*MockProductRepository)
		expectedError error
	}{
		{
			name:        "успешное получение",
			receptionID: uuid.New(),
			setupMocks: func(repo *MockProductRepository) {
				repo.On("GetByReceptionID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return([]*product.Product{}, nil)
			},
			expectedError: nil,
		},
		{
			name:        "ошибка получения",
			receptionID: uuid.New(),
			setupMocks: func(repo *MockProductRepository) {
				repo.On("GetByReceptionID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return([]*product.Product(nil), errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(MockProductRepository)
			tt.setupMocks(productRepo)

			service := New(productRepo, nil, nil)
			_, err := service.GetByReceptionID(context.Background(), tt.receptionID)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			productRepo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	tests := []struct {
		name          string
		offset        int
		limit         int
		setupMocks    func(*MockProductRepository)
		expectedError error
	}{
		{
			name:   "успешное получение списка",
			offset: 0,
			limit:  10,
			setupMocks: func(repo *MockProductRepository) {
				repo.On("List", mock.Anything, 0, 10).Return([]*product.Product{}, nil)
			},
			expectedError: nil,
		},
		{
			name:   "ошибка получения списка",
			offset: 0,
			limit:  10,
			setupMocks: func(repo *MockProductRepository) {
				repo.On("List", mock.Anything, 0, 10).Return([]*product.Product(nil), errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(MockProductRepository)
			tt.setupMocks(productRepo)

			service := New(productRepo, nil, nil)
			_, err := service.List(context.Background(), tt.offset, tt.limit)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			productRepo.AssertExpectations(t)
		})
	}
}

func TestService_AddProduct(t *testing.T) {
	tests := []struct {
		name          string
		receptionID   uuid.UUID
		productType   product.Type
		setupMocks    func(*MockProductRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name:        "успешное добавление",
			receptionID: uuid.New(),
			productType: product.TypeElectronics,
			setupMocks: func(productRepo *MockProductRepository, tx *MockTransactionManager) {
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
				productRepo.On("Create", mock.Anything, mock.AnythingOfType("*product.Product")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "ошибка добавления",
			receptionID: uuid.New(),
			productType: product.TypeElectronics,
			setupMocks: func(productRepo *MockProductRepository, tx *MockTransactionManager) {
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(errors.New("create error"))
				productRepo.On("Create", mock.Anything, mock.AnythingOfType("*product.Product")).Return(errors.New("create error"))
			},
			expectedError: errors.New("create error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(MockProductRepository)
			tx := new(MockTransactionManager)
			tt.setupMocks(productRepo, tx)

			service := New(productRepo, nil, tx)
			_, err := service.AddProduct(context.Background(), tt.receptionID, tt.productType)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			productRepo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}

func TestService_AddProducts(t *testing.T) {
	tests := []struct {
		name          string
		receptionID   uuid.UUID
		types         []product.Type
		setupMocks    func(*MockProductRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name:        "успешное добавление",
			receptionID: uuid.New(),
			types:       []product.Type{product.TypeElectronics, product.TypeClothing},
			setupMocks: func(productRepo *MockProductRepository, tx *MockTransactionManager) {
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
				productRepo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*product.Product")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "ошибка добавления",
			receptionID: uuid.New(),
			types:       []product.Type{product.TypeElectronics},
			setupMocks: func(productRepo *MockProductRepository, tx *MockTransactionManager) {
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(errors.New("create error"))
				productRepo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*product.Product")).Return(errors.New("create error"))
			},
			expectedError: errors.New("create error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(MockProductRepository)
			tx := new(MockTransactionManager)
			tt.setupMocks(productRepo, tx)

			service := New(productRepo, nil, tx)
			_, err := service.AddProducts(context.Background(), tt.receptionID, tt.types)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			productRepo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}

func TestService_DeleteLastProduct(t *testing.T) {
	tests := []struct {
		name          string
		receptionID   uuid.UUID
		setupMocks    func(*MockProductRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name:        "успешное удаление",
			receptionID: uuid.New(),
			setupMocks: func(productRepo *MockProductRepository, tx *MockTransactionManager) {
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
				productRepo.On("DeleteLast", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "ошибка удаления",
			receptionID: uuid.New(),
			setupMocks: func(productRepo *MockProductRepository, tx *MockTransactionManager) {
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(errors.New("delete error"))
				productRepo.On("DeleteLast", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(errors.New("delete error"))
			},
			expectedError: errors.New("delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(MockProductRepository)
			tx := new(MockTransactionManager)
			tt.setupMocks(productRepo, tx)

			service := New(productRepo, nil, tx)
			err := service.DeleteLastProduct(context.Background(), tt.receptionID)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			productRepo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name          string
		receptionID   uuid.UUID
		productType   product.Type
		setupMocks    func(*MockProductRepository, *MockReceptionRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name:        "успешное создание",
			receptionID: uuid.New(),
			productType: product.TypeElectronics,
			setupMocks: func(productRepo *MockProductRepository, receptionRepo *MockReceptionRepository, tx *MockTransactionManager) {
				receptionRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&reception.Reception{Status: reception.StatusInProgress}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
				productRepo.On("Create", mock.Anything, mock.AnythingOfType("*product.Product")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "приемка не найдена",
			receptionID: uuid.New(),
			productType: product.TypeElectronics,
			setupMocks: func(productRepo *MockProductRepository, receptionRepo *MockReceptionRepository, tx *MockTransactionManager) {
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
			productType: product.TypeElectronics,
			setupMocks: func(productRepo *MockProductRepository, receptionRepo *MockReceptionRepository, tx *MockTransactionManager) {
				receptionRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&reception.Reception{Status: reception.StatusClose}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrReceptionAlreadyClose)
			},
			expectedError: ErrReceptionAlreadyClose,
		},
		{
			name:        "некорректный тип товара",
			receptionID: uuid.New(),
			productType: product.Type("invalid"),
			setupMocks: func(productRepo *MockProductRepository, receptionRepo *MockReceptionRepository, tx *MockTransactionManager) {
				receptionRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&reception.Reception{Status: reception.StatusInProgress}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrInvalidProductType)
			},
			expectedError: ErrInvalidProductType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(MockProductRepository)
			receptionRepo := new(MockReceptionRepository)
			tx := new(MockTransactionManager)
			tt.setupMocks(productRepo, receptionRepo, tx)

			service := New(productRepo, receptionRepo, tx)
			_, err := service.Create(context.Background(), tt.receptionID, tt.productType)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			productRepo.AssertExpectations(t)
			receptionRepo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}
