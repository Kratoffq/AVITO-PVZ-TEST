package pvz

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/avito/pvz/internal/domain/pvz"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/avito/pvz/internal/domain/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPVZRepository мок репозитория ПВЗ
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

func (m *MockPVZRepository) GetByCity(ctx context.Context, city string) (*pvz.PVZ, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) Update(ctx context.Context, pvz *pvz.PVZ) error {
	args := m.Called(ctx, pvz)
	return args.Error(0)
}

func (m *MockPVZRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPVZRepository) GetAll(ctx context.Context) ([]*pvz.PVZ, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) List(ctx context.Context, offset, limit int) ([]*pvz.PVZ, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) GetWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*pvz.PVZWithReceptions, error) {
	args := m.Called(ctx, startDate, endDate, page, limit)
	return args.Get(0).([]*pvz.PVZWithReceptions), args.Error(1)
}

// MockUserRepository мок репозитория пользователей
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*user.User, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*user.User), args.Error(1)
}

// MockTransactionManager мок менеджера транзакций
type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

// MockAuditLog мок журнала аудита
type MockAuditLog struct {
	mock.Mock
}

func (m *MockAuditLog) LogPVZCreation(ctx context.Context, pvzID, userID uuid.UUID) error {
	args := m.Called(ctx, pvzID, userID)
	return args.Error(0)
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name        string
		city        string
		userID      uuid.UUID
		setupMocks  func(*MockPVZRepository, *MockUserRepository, *MockTransactionManager, *MockAuditLog)
		expectedErr error
	}{
		{
			name:   "успешное создание ПВЗ",
			city:   "Москва",
			userID: uuid.New(),
			setupMocks: func(pvzRepo *MockPVZRepository, userRepo *MockUserRepository, txManager *MockTransactionManager, auditLog *MockAuditLog) {
				user := &user.User{Role: user.RoleAdmin}
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
				pvzRepo.On("GetByCity", mock.Anything, "Москва").Return(nil, pvz.ErrNotFound)
				pvzRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				auditLog.On("LogPVZCreation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				txManager.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "неверный город",
			city:   "",
			userID: uuid.New(),
			setupMocks: func(pvzRepo *MockPVZRepository, userRepo *MockUserRepository, txManager *MockTransactionManager, auditLog *MockAuditLog) {
				// Моки не нужны, так как валидация происходит до их вызова
			},
			expectedErr: ErrInvalidCity,
		},
		{
			name:   "нет прав доступа",
			city:   "Москва",
			userID: uuid.New(),
			setupMocks: func(pvzRepo *MockPVZRepository, userRepo *MockUserRepository, txManager *MockTransactionManager, auditLog *MockAuditLog) {
				user := &user.User{Role: user.RoleUser}
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
			},
			expectedErr: ErrAccessDenied,
		},
		{
			name:   "ПВЗ уже существует",
			city:   "Москва",
			userID: uuid.New(),
			setupMocks: func(pvzRepo *MockPVZRepository, userRepo *MockUserRepository, txManager *MockTransactionManager, auditLog *MockAuditLog) {
				user := &user.User{Role: user.RoleAdmin}
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
				existingPVZ := &pvz.PVZ{City: "Москва"}
				pvzRepo.On("GetByCity", mock.Anything, "Москва").Return(existingPVZ, nil)
				txManager.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrPVZAlreadyExists)
			},
			expectedErr: ErrPVZAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(MockPVZRepository)
			userRepo := new(MockUserRepository)
			txManager := new(MockTransactionManager)
			auditLog := new(MockAuditLog)

			tt.setupMocks(pvzRepo, userRepo, txManager, auditLog)

			service := New(pvzRepo, userRepo, txManager, auditLog, nil)
			result, err := service.Create(context.Background(), tt.city, tt.userID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.city, result.City)
				assert.NotEqual(t, uuid.Nil, result.ID)
			}

			pvzRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
			auditLog.AssertExpectations(t)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	tests := []struct {
		name        string
		id          uuid.UUID
		setupMocks  func(*MockPVZRepository)
		expectedPVZ *pvz.PVZ
		expectedErr error
	}{
		{
			name: "успешное получение ПВЗ",
			id:   uuid.New(),
			setupMocks: func(pvzRepo *MockPVZRepository) {
				expectedPVZ := &pvz.PVZ{
					ID:   uuid.New(),
					City: "Москва",
				}
				pvzRepo.On("GetByID", mock.Anything, mock.Anything).Return(expectedPVZ, nil)
			},
			expectedPVZ: &pvz.PVZ{
				ID:   uuid.New(),
				City: "Москва",
			},
			expectedErr: nil,
		},
		{
			name: "ПВЗ не найден",
			id:   uuid.New(),
			setupMocks: func(pvzRepo *MockPVZRepository) {
				pvzRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, pvz.ErrNotFound)
			},
			expectedPVZ: nil,
			expectedErr: ErrPVZNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(MockPVZRepository)
			tt.setupMocks(pvzRepo)

			service := New(pvzRepo, nil, nil, nil, nil)
			result, err := service.GetByID(context.Background(), tt.id)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedPVZ.City, result.City)
			}

			pvzRepo.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name        string
		pvz         *pvz.PVZ
		moderatorID uuid.UUID
		setupMocks  func(*MockPVZRepository, *MockUserRepository, *MockTransactionManager)
		expectedErr error
	}{
		{
			name: "успешное обновление ПВЗ",
			pvz: &pvz.PVZ{
				ID:   uuid.New(),
				City: "Санкт-Петербург",
			},
			moderatorID: uuid.New(),
			setupMocks: func(pvzRepo *MockPVZRepository, userRepo *MockUserRepository, txManager *MockTransactionManager) {
				moderator := &user.User{Role: user.RoleAdmin}
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(moderator, nil)
				pvzRepo.On("GetByID", mock.Anything, mock.Anything).Return(&pvz.PVZ{}, nil)
				pvzRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
				txManager.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "нет прав доступа",
			pvz: &pvz.PVZ{
				ID:   uuid.New(),
				City: "Санкт-Петербург",
			},
			moderatorID: uuid.New(),
			setupMocks: func(pvzRepo *MockPVZRepository, userRepo *MockUserRepository, txManager *MockTransactionManager) {
				moderator := &user.User{Role: user.RoleUser}
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(moderator, nil)
			},
			expectedErr: ErrAccessDenied,
		},
		{
			name: "неверные данные ПВЗ",
			pvz: &pvz.PVZ{
				ID:   uuid.New(),
				City: "",
			},
			moderatorID: uuid.New(),
			setupMocks: func(pvzRepo *MockPVZRepository, userRepo *MockUserRepository, txManager *MockTransactionManager) {
				moderator := &user.User{Role: user.RoleAdmin}
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(moderator, nil)
			},
			expectedErr: ErrInvalidCity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(MockPVZRepository)
			userRepo := new(MockUserRepository)
			txManager := new(MockTransactionManager)

			tt.setupMocks(pvzRepo, userRepo, txManager)

			service := New(pvzRepo, userRepo, txManager, nil, nil)
			err := service.Update(context.Background(), tt.pvz, tt.moderatorID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			pvzRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		id          uuid.UUID
		moderatorID uuid.UUID
		setupMocks  func(*MockPVZRepository, *MockUserRepository, *MockTransactionManager)
		expectedErr error
	}{
		{
			name:        "успешное удаление ПВЗ",
			id:          uuid.New(),
			moderatorID: uuid.New(),
			setupMocks: func(pvzRepo *MockPVZRepository, userRepo *MockUserRepository, txManager *MockTransactionManager) {
				moderator := &user.User{Role: user.RoleAdmin}
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(moderator, nil)
				pvzRepo.On("GetByID", mock.Anything, mock.Anything).Return(&pvz.PVZ{}, nil)
				pvzRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)
				txManager.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "нет прав доступа",
			id:          uuid.New(),
			moderatorID: uuid.New(),
			setupMocks: func(pvzRepo *MockPVZRepository, userRepo *MockUserRepository, txManager *MockTransactionManager) {
				moderator := &user.User{Role: user.RoleUser}
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(moderator, nil)
			},
			expectedErr: ErrAccessDenied,
		},
		{
			name:        "ПВЗ не найден",
			id:          uuid.New(),
			moderatorID: uuid.New(),
			setupMocks: func(pvzRepo *MockPVZRepository, userRepo *MockUserRepository, txManager *MockTransactionManager) {
				moderator := &user.User{Role: user.RoleAdmin}
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(moderator, nil)
				pvzRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, pvz.ErrNotFound)
				txManager.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrPVZNotFound)
			},
			expectedErr: ErrPVZNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(MockPVZRepository)
			userRepo := new(MockUserRepository)
			txManager := new(MockTransactionManager)

			tt.setupMocks(pvzRepo, userRepo, txManager)

			service := New(pvzRepo, userRepo, txManager, nil, nil)
			err := service.Delete(context.Background(), tt.id, tt.moderatorID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			pvzRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}

func TestService_GetAll(t *testing.T) {
	tests := []struct {
		name         string
		setupMocks   func(*MockPVZRepository)
		expectedPVZs []*pvz.PVZ
		expectedErr  error
	}{
		{
			name: "успешное получение всех ПВЗ",
			setupMocks: func(pvzRepo *MockPVZRepository) {
				expectedPVZs := []*pvz.PVZ{
					{ID: uuid.New(), City: "Москва"},
					{ID: uuid.New(), City: "Санкт-Петербург"},
				}
				pvzRepo.On("GetAll", mock.Anything).Return(expectedPVZs, nil)
			},
			expectedPVZs: []*pvz.PVZ{
				{ID: uuid.New(), City: "Москва"},
				{ID: uuid.New(), City: "Санкт-Петербург"},
			},
			expectedErr: nil,
		},
		{
			name: "ошибка при получении ПВЗ",
			setupMocks: func(pvzRepo *MockPVZRepository) {
				pvzRepo.On("GetAll", mock.Anything).Return(nil, assert.AnError)
			},
			expectedPVZs: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(MockPVZRepository)
			tt.setupMocks(pvzRepo)

			service := New(pvzRepo, nil, nil, nil, nil)
			result, err := service.GetAll(context.Background())

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedPVZs), len(result))
				for i, pvz := range result {
					assert.Equal(t, tt.expectedPVZs[i].City, pvz.City)
				}
			}

			pvzRepo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	tests := []struct {
		name         string
		offset       int
		limit        int
		setupMocks   func(*MockPVZRepository)
		expectedPVZs []*pvz.PVZ
		expectedErr  error
	}{
		{
			name:   "успешное получение списка ПВЗ",
			offset: 0,
			limit:  10,
			setupMocks: func(pvzRepo *MockPVZRepository) {
				expectedPVZs := []*pvz.PVZ{
					{ID: uuid.New(), City: "Москва"},
					{ID: uuid.New(), City: "Санкт-Петербург"},
				}
				pvzRepo.On("List", mock.Anything, 0, 10).Return(expectedPVZs, nil)
			},
			expectedPVZs: []*pvz.PVZ{
				{ID: uuid.New(), City: "Москва"},
				{ID: uuid.New(), City: "Санкт-Петербург"},
			},
			expectedErr: nil,
		},
		{
			name:   "ошибка при получении списка ПВЗ",
			offset: 0,
			limit:  10,
			setupMocks: func(pvzRepo *MockPVZRepository) {
				pvzRepo.On("List", mock.Anything, 0, 10).Return(nil, assert.AnError)
			},
			expectedPVZs: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(MockPVZRepository)
			tt.setupMocks(pvzRepo)

			service := New(pvzRepo, nil, nil, nil, nil)
			result, err := service.List(context.Background(), tt.offset, tt.limit)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedPVZs), len(result))
				for i, pvz := range result {
					assert.Equal(t, tt.expectedPVZs[i].City, pvz.City)
				}
			}

			pvzRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetWithReceptions(t *testing.T) {
	tests := []struct {
		name        string
		startDate   time.Time
		endDate     time.Time
		page        int
		limit       int
		setupMocks  func(*MockPVZRepository)
		expectedErr error
	}{
		{
			name:      "успешное получение ПВЗ с приемками",
			startDate: time.Now().Add(-24 * time.Hour),
			endDate:   time.Now(),
			page:      1,
			limit:     10,
			setupMocks: func(pvzRepo *MockPVZRepository) {
				pvzs := []*pvz.PVZWithReceptions{
					{
						PVZ: &pvz.PVZ{
							ID:   uuid.New(),
							City: "Москва",
						},
						Receptions: []*pvz.ReceptionWithProducts{
							{
								Reception: &reception.Reception{
									ID:     uuid.New(),
									Status: "open",
								},
								Products: []*product.Product{},
							},
						},
					},
				}
				pvzRepo.On("GetWithReceptions", mock.Anything, mock.Anything, mock.Anything, 1, 10).Return(pvzs, nil)
			},
			expectedErr: nil,
		},
		{
			name:      "неверные параметры пагинации",
			startDate: time.Now().Add(-24 * time.Hour),
			endDate:   time.Now(),
			page:      -1,
			limit:     0,
			setupMocks: func(pvzRepo *MockPVZRepository) {
				// Моки не нужны, так как валидация происходит до их вызова
			},
			expectedErr: errors.New("invalid pagination parameters"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(MockPVZRepository)
			tt.setupMocks(pvzRepo)

			service := New(pvzRepo, nil, nil, nil, nil)
			result, err := service.GetWithReceptions(context.Background(), tt.startDate, tt.endDate, tt.page, tt.limit)

			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result)
				for _, pvz := range result {
					assert.NotNil(t, pvz.PVZ)
					assert.NotEmpty(t, pvz.PVZ.City)
					assert.NotEqual(t, uuid.Nil, pvz.PVZ.ID)
				}
			}

			pvzRepo.AssertExpectations(t)
		})
	}
}

func TestValidateCity(t *testing.T) {
	tests := []struct {
		name          string
		city          string
		expectedError error
	}{
		{
			name:          "валидный город",
			city:          "Москва",
			expectedError: nil,
		},
		{
			name:          "пустой город",
			city:          "",
			expectedError: ErrInvalidCity,
		},
		{
			name:          "слишком короткое название",
			city:          "a",
			expectedError: ErrInvalidCity,
		},
		{
			name:          "слишком длинное название",
			city:          strings.Repeat("А", 101),
			expectedError: ErrInvalidCity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCity(tt.city)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
