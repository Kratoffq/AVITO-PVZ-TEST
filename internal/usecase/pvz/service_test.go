package pvz

import (
	"context"
	"testing"
	"time"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/avito/pvz/internal/domain/pvz"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPVZRepository мок для репозитория ПВЗ
type MockPVZRepository struct {
	mock.Mock
}

func (m *MockPVZRepository) Create(ctx context.Context, pvz *pvz.PVZ) error {
	args := m.Called(ctx, pvz)
	return args.Error(0)
}

func (m *MockPVZRepository) GetByID(ctx context.Context, id uuid.UUID) (*pvz.PVZ, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) GetWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*pvz.PVZWithReceptions, error) {
	args := m.Called(ctx, startDate, endDate, page, limit)
	return args.Get(0).([]*pvz.PVZWithReceptions), args.Error(1)
}

func (m *MockPVZRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPVZRepository) GetAll(ctx context.Context) ([]*pvz.PVZ, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) GetByCity(ctx context.Context, city string) (*pvz.PVZ, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) List(ctx context.Context, offset, limit int) ([]*pvz.PVZ, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) Update(ctx context.Context, pvz *pvz.PVZ) error {
	args := m.Called(ctx, pvz)
	return args.Error(0)
}

// MockReceptionRepository мок для репозитория приемок
type MockReceptionRepository struct {
	mock.Mock
}

func (m *MockReceptionRepository) Create(ctx context.Context, reception *reception.Reception) error {
	args := m.Called(ctx, reception)
	return args.Error(0)
}

func (m *MockReceptionRepository) GetOpenByPVZID(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
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

func (m *MockReceptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*reception.Reception, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception.Reception), args.Error(1)
}

func (m *MockReceptionRepository) GetLastOpen(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
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

func (m *MockReceptionRepository) List(ctx context.Context, offset, limit int) ([]*reception.Reception, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*reception.Reception), args.Error(1)
}

func (m *MockReceptionRepository) Update(ctx context.Context, reception *reception.Reception) error {
	args := m.Called(ctx, reception)
	return args.Error(0)
}

// MockTransactionManager мок для менеджера транзакций
type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func TestCreatePVZ(t *testing.T) {
	tests := []struct {
		name          string
		req           *CreatePVZRequest
		mockSetup     func(*MockPVZRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name: "успешное создание ПВЗ",
			req: &CreatePVZRequest{
				City: "Москва",
			},
			mockSetup: func(pvzRepo *MockPVZRepository, txManager *MockTransactionManager) {
				txManager.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(context.Background())
					}).
					Return(nil)

				pvzRepo.On("Create", mock.Anything, mock.AnythingOfType("*pvz.PVZ")).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "ошибка при создании ПВЗ",
			req: &CreatePVZRequest{
				City: "Москва",
			},
			mockSetup: func(pvzRepo *MockPVZRepository, txManager *MockTransactionManager) {
				txManager.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(context.Background())
					}).
					Return(assert.AnError)

				pvzRepo.On("Create", mock.Anything, mock.AnythingOfType("*pvz.PVZ")).
					Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(MockPVZRepository)
			receptionRepo := new(MockReceptionRepository)
			txManager := new(MockTransactionManager)

			tt.mockSetup(pvzRepo, txManager)

			useCase := New(pvzRepo, receptionRepo, txManager)
			resp, err := useCase.CreatePVZ(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.req.City, resp.City)
			}

			pvzRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}

func TestCreateReception(t *testing.T) {
	tests := []struct {
		name          string
		pvzID         uuid.UUID
		mockSetup     func(*MockPVZRepository, *MockReceptionRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name:  "успешное создание приемки",
			pvzID: uuid.New(),
			mockSetup: func(pvzRepo *MockPVZRepository, receptionRepo *MockReceptionRepository, txManager *MockTransactionManager) {
				txManager.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(context.Background())
					}).
					Return(nil)

				pvzRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&pvz.PVZ{}, nil)

				receptionRepo.On("GetOpenByPVZID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, reception.ErrNoOpenReception)

				receptionRepo.On("Create", mock.Anything, mock.AnythingOfType("*reception.Reception")).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "ошибка при существующей открытой приемке",
			pvzID: uuid.New(),
			mockSetup: func(pvzRepo *MockPVZRepository, receptionRepo *MockReceptionRepository, txManager *MockTransactionManager) {
				txManager.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(context.Background())
					}).
					Return(reception.ErrReceptionAlreadyOpen)

				pvzRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&pvz.PVZ{}, nil)

				receptionRepo.On("GetOpenByPVZID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&reception.Reception{}, nil)
			},
			expectedError: reception.ErrReceptionAlreadyOpen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(MockPVZRepository)
			receptionRepo := new(MockReceptionRepository)
			txManager := new(MockTransactionManager)

			tt.mockSetup(pvzRepo, receptionRepo, txManager)

			useCase := New(pvzRepo, receptionRepo, txManager)
			resp, err := useCase.CreateReception(context.Background(), tt.pvzID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}

			pvzRepo.AssertExpectations(t)
			receptionRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}

func TestGetPVZWithReceptions(t *testing.T) {
	tests := []struct {
		name          string
		req           *GetPVZWithReceptionsRequest
		mockSetup     func(*MockPVZRepository)
		expectedError error
	}{
		{
			name: "успешное получение ПВЗ с приемками",
			req: &GetPVZWithReceptionsRequest{
				StartDate: time.Now(),
				EndDate:   time.Now().Add(24 * time.Hour),
				Page:      1,
				Limit:     10,
			},
			mockSetup: func(pvzRepo *MockPVZRepository) {
				pvzRepo.On("GetWithReceptions", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 1, 10).
					Return([]*pvz.PVZWithReceptions{
						{
							PVZ: &pvz.PVZ{
								ID:        uuid.New(),
								CreatedAt: time.Now(),
								City:      "Москва",
							},
							Receptions: []*pvz.ReceptionWithProducts{},
						},
					}, nil)
			},
			expectedError: nil,
		},
		{
			name: "ошибка при получении ПВЗ с приемками",
			req: &GetPVZWithReceptionsRequest{
				StartDate: time.Now(),
				EndDate:   time.Now().Add(24 * time.Hour),
				Page:      1,
				Limit:     10,
			},
			mockSetup: func(pvzRepo *MockPVZRepository) {
				pvzRepo.On("GetWithReceptions", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 1, 10).
					Return([]*pvz.PVZWithReceptions{}, assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(MockPVZRepository)
			receptionRepo := new(MockReceptionRepository)
			txManager := new(MockTransactionManager)

			tt.mockSetup(pvzRepo)

			useCase := New(pvzRepo, receptionRepo, txManager)
			resp, err := useCase.GetPVZWithReceptions(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Items)
				assert.Len(t, resp.Items, 1)
				assert.Equal(t, "Москва", resp.Items[0].City)
			}

			pvzRepo.AssertExpectations(t)
		})
	}
}

func TestCreatePVZ_Validation(t *testing.T) {
	tests := []struct {
		name          string
		req           *CreatePVZRequest
		mockSetup     func(*MockPVZRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name: "пустой город",
			req: &CreatePVZRequest{
				City: "",
			},
			mockSetup: func(pvzRepo *MockPVZRepository, txManager *MockTransactionManager) {
				txManager.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(pvz.ErrInvalidCity)
			},
			expectedError: pvz.ErrInvalidCity,
		},
		{
			name: "ошибка транзакции",
			req: &CreatePVZRequest{
				City: "Москва",
			},
			mockSetup: func(pvzRepo *MockPVZRepository, txManager *MockTransactionManager) {
				txManager.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(MockPVZRepository)
			receptionRepo := new(MockReceptionRepository)
			txManager := new(MockTransactionManager)

			tt.mockSetup(pvzRepo, txManager)

			useCase := New(pvzRepo, receptionRepo, txManager)
			resp, err := useCase.CreatePVZ(context.Background(), tt.req)

			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err)
			assert.Nil(t, resp)

			pvzRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}

func TestGetPVZWithReceptions_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		req           *GetPVZWithReceptionsRequest
		mockSetup     func(*MockPVZRepository)
		expectedError error
	}{
		{
			name: "неверный период дат",
			req: &GetPVZWithReceptionsRequest{
				StartDate: time.Now().Add(24 * time.Hour),
				EndDate:   time.Now(),
				Page:      1,
				Limit:     10,
			},
			mockSetup: func(pvzRepo *MockPVZRepository) {
				pvzRepo.On("GetWithReceptions", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 1, 10).
					Return(nil, pvz.ErrInvalidDateRange)
			},
			expectedError: pvz.ErrInvalidDateRange,
		},
		{
			name: "неверная пагинация",
			req: &GetPVZWithReceptionsRequest{
				StartDate: time.Now(),
				EndDate:   time.Now().Add(24 * time.Hour),
				Page:      0,
				Limit:     10,
			},
			mockSetup: func(pvzRepo *MockPVZRepository) {
				pvzRepo.On("GetWithReceptions", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 0, 10).
					Return(nil, pvz.ErrInvalidPagination)
			},
			expectedError: pvz.ErrInvalidPagination,
		},
		{
			name: "пустой результат",
			req: &GetPVZWithReceptionsRequest{
				StartDate: time.Now(),
				EndDate:   time.Now().Add(24 * time.Hour),
				Page:      1,
				Limit:     10,
			},
			mockSetup: func(pvzRepo *MockPVZRepository) {
				pvzRepo.On("GetWithReceptions", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 1, 10).
					Return([]*pvz.PVZWithReceptions{}, nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(MockPVZRepository)
			receptionRepo := new(MockReceptionRepository)
			txManager := new(MockTransactionManager)

			tt.mockSetup(pvzRepo)

			useCase := New(pvzRepo, receptionRepo, txManager)
			resp, err := useCase.GetPVZWithReceptions(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Empty(t, resp.Items)
			}

			pvzRepo.AssertExpectations(t)
		})
	}
}

func TestCreateReception_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		pvzID         uuid.UUID
		mockSetup     func(*MockPVZRepository, *MockReceptionRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name:  "ПВЗ не существует",
			pvzID: uuid.New(),
			mockSetup: func(pvzRepo *MockPVZRepository, receptionRepo *MockReceptionRepository, txManager *MockTransactionManager) {
				txManager.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(context.Background())
					}).
					Return(pvz.ErrNotFound)

				pvzRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, pvz.ErrNotFound)
			},
			expectedError: pvz.ErrNotFound,
		},
		{
			name:  "ошибка при создании приемки",
			pvzID: uuid.New(),
			mockSetup: func(pvzRepo *MockPVZRepository, receptionRepo *MockReceptionRepository, txManager *MockTransactionManager) {
				txManager.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(context.Background())
					}).
					Return(assert.AnError)

				pvzRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&pvz.PVZ{}, nil)

				receptionRepo.On("GetOpenByPVZID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, reception.ErrNoOpenReception)

				receptionRepo.On("Create", mock.Anything, mock.AnythingOfType("*reception.Reception")).
					Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(MockPVZRepository)
			receptionRepo := new(MockReceptionRepository)
			txManager := new(MockTransactionManager)

			tt.mockSetup(pvzRepo, receptionRepo, txManager)

			useCase := New(pvzRepo, receptionRepo, txManager)
			resp, err := useCase.CreateReception(context.Background(), tt.pvzID)

			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err)
			assert.Nil(t, resp)

			pvzRepo.AssertExpectations(t)
			receptionRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}
