package impl

import (
	"context"
	"testing"
	"time"

	"github.com/avito/pvz/internal/config"
	"github.com/avito/pvz/internal/models"
	"github.com/avito/pvz/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreatePVZ(ctx context.Context, pvz *models.PVZ) error {
	args := m.Called(ctx, pvz)
	return args.Error(0)
}

func (m *MockRepository) GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PVZ), args.Error(1)
}

func (m *MockRepository) GetPVZsWithReceptions(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*repository.PVZWithReceptions, error) {
	args := m.Called(ctx, startTime, endTime, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.PVZWithReceptions), args.Error(1)
}

func (m *MockRepository) GetLastOpenReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *MockRepository) GetReceptionByID(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *MockRepository) CreateReception(ctx context.Context, reception *models.Reception) error {
	args := m.Called(ctx, reception)
	return args.Error(0)
}

func (m *MockRepository) CloseReception(ctx context.Context, receptionID uuid.UUID) error {
	args := m.Called(ctx, receptionID)
	return args.Error(0)
}

func (m *MockRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockRepository) CreateProducts(ctx context.Context, products []*models.Product) error {
	args := m.Called(ctx, products)
	return args.Error(0)
}

func (m *MockRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) GetLastProduct(ctx context.Context, receptionID uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, receptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockRepository) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

func TestService_CreatePVZ(t *testing.T) {
	mockRepo := new(MockRepository)
	cfg := &config.Config{}
	service := NewService(mockRepo, cfg)

	ctx := context.Background()
	city := "Москва"

	mockRepo.On("CreatePVZ", ctx, mock.AnythingOfType("*models.PVZ")).Return(nil)

	pvz, err := service.CreatePVZ(ctx, city)
	assert.NoError(t, err)
	assert.NotNil(t, pvz)
	assert.Equal(t, city, pvz.City)
	mockRepo.AssertExpectations(t)
}

func TestService_CreateReception(t *testing.T) {
	mockRepo := new(MockRepository)
	cfg := &config.Config{}
	service := NewService(mockRepo, cfg)

	ctx := context.Background()
	pvzID := uuid.New()

	mockRepo.On("GetPVZByID", ctx, pvzID).Return(&models.PVZ{ID: pvzID}, nil)
	mockRepo.On("GetLastOpenReception", ctx, pvzID).Return(nil, nil)
	mockRepo.On("CreateReception", ctx, mock.AnythingOfType("*models.Reception")).Return(nil)

	reception, err := service.CreateReception(ctx, pvzID)
	assert.NoError(t, err)
	assert.NotNil(t, reception)
	assert.Equal(t, pvzID, reception.PVZID)
	mockRepo.AssertExpectations(t)
}

func TestService_AddProduct(t *testing.T) {
	mockRepo := new(MockRepository)
	cfg := &config.Config{}
	service := NewService(mockRepo, cfg)

	ctx := context.Background()
	pvzID := uuid.New()
	productType := models.TypeElectronics

	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.StatusInProgress,
	}

	mockRepo.On("GetLastOpenReception", ctx, pvzID).Return(reception, nil)
	mockRepo.On("CreateProduct", ctx, mock.AnythingOfType("*models.Product")).Return(nil)

	product, err := service.AddProduct(ctx, pvzID, productType)
	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, productType, product.Type)
	assert.Equal(t, reception.ID, product.ReceptionID)
	mockRepo.AssertExpectations(t)
}

func TestService_CloseReception(t *testing.T) {
	mockRepo := new(MockRepository)
	cfg := &config.Config{}
	service := NewService(mockRepo, cfg)

	ctx := context.Background()
	pvzID := uuid.New()
	receptionID := uuid.New()

	reception := &models.Reception{
		ID:       receptionID,
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.StatusInProgress,
	}

	mockRepo.On("GetLastOpenReception", ctx, pvzID).Return(reception, nil)
	mockRepo.On("CloseReception", ctx, receptionID).Return(nil)

	closedReception, err := service.CloseReception(ctx, pvzID)
	assert.NoError(t, err)
	assert.NotNil(t, closedReception)
	assert.Equal(t, models.StatusClose, closedReception.Status)
	mockRepo.AssertExpectations(t)
}

func TestService_DeleteLastProduct(t *testing.T) {
	mockRepo := new(MockRepository)
	cfg := &config.Config{}
	service := NewService(mockRepo, cfg)

	ctx := context.Background()
	pvzID := uuid.New()
	receptionID := uuid.New()
	productID := uuid.New()

	reception := &models.Reception{
		ID:       receptionID,
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.StatusInProgress,
	}

	product := &models.Product{
		ID:          productID,
		DateTime:    time.Now(),
		Type:        models.TypeElectronics,
		ReceptionID: receptionID,
	}

	mockRepo.On("GetLastOpenReception", ctx, pvzID).Return(reception, nil)
	mockRepo.On("GetLastProduct", ctx, receptionID).Return(product, nil)
	mockRepo.On("DeleteProduct", ctx, productID).Return(nil)

	err := service.DeleteLastProduct(ctx, pvzID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
