package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/avito/pvz/internal/config"
	"github.com/avito/pvz/internal/models"
	"github.com/avito/pvz/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Определяем контекст сервиса
type serviceContext struct {
	context.Context
}

func (c serviceContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (c serviceContext) Done() <-chan struct{} {
	return nil
}

func (c serviceContext) Err() error {
	return nil
}

func (c serviceContext) Value(key interface{}) interface{} {
	return nil
}

// Определяем структуру для PVZ с рецепциями
type pvzWithReceptions struct {
	PVZ        *models.PVZ
	Receptions []*models.Reception
}

type MockService struct {
	mock.Mock
}

func (m *MockService) RegisterUser(ctx context.Context, email, password string, role models.UserRole) (*models.User, error) {
	args := m.Called(ctx, email, password, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockService) LoginUser(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockService) DummyLogin(ctx context.Context, role models.UserRole) (string, error) {
	args := m.Called(ctx, role)
	return args.String(0), args.Error(1)
}

func (m *MockService) CreatePVZ(ctx context.Context, city string) (*models.PVZ, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PVZ), args.Error(1)
}

func (m *MockService) GetPVZsWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*repository.PVZWithReceptions, error) {
	args := m.Called(ctx, startDate, endDate, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.PVZWithReceptions), args.Error(1)
}

func (m *MockService) CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *MockService) CloseReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *MockService) AddProduct(ctx context.Context, pvzID uuid.UUID, productType models.ProductType) (*models.Product, error) {
	args := m.Called(ctx, pvzID, productType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockService) AddProducts(ctx context.Context, pvzID uuid.UUID, productTypes []models.ProductType) ([]*models.Product, error) {
	args := m.Called(ctx, pvzID, productTypes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	args := m.Called(ctx, pvzID)
	return args.Error(0)
}

func TestHandler_Register(t *testing.T) {
	mockService := new(MockService)
	cfg := &config.Config{}
	h := NewHandler(mockService, cfg)

	router := gin.Default()
	router.POST("/auth/register", h.register)

	reqBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
		"role":     "employee",
	}
	jsonBody, _ := json.Marshal(reqBody)

	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Role:      models.RoleEmployee,
		CreatedAt: time.Now(),
	}

	mockService.On("RegisterUser", mock.Anything, "test@example.com", "password123", models.RoleEmployee).Return(user, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_Login(t *testing.T) {
	mockService := new(MockService)
	cfg := &config.Config{}
	h := NewHandler(mockService, cfg)

	router := gin.Default()
	router.POST("/auth/login", h.login)

	reqBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("LoginUser", mock.Anything, "test@example.com", "password123").Return("token123", nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_CreatePVZ(t *testing.T) {
	mockService := new(MockService)
	cfg := &config.Config{}
	h := NewHandler(mockService, cfg)

	router := gin.Default()
	router.POST("/api/pvz", h.authMiddleware, h.createPVZ)

	reqBody := map[string]interface{}{
		"city": "Москва",
	}
	jsonBody, _ := json.Marshal(reqBody)

	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             "Москва",
	}

	mockService.On("CreatePVZ", mock.Anything, "Москва").Return(pvz, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pvz", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_CreateReception(t *testing.T) {
	mockService := new(MockService)
	cfg := &config.Config{}
	h := NewHandler(mockService, cfg)

	router := gin.Default()
	router.POST("/api/receptions", h.authMiddleware, h.createReception)

	pvzID := uuid.New()
	reqBody := map[string]interface{}{
		"pvz_id": pvzID.String(),
	}
	jsonBody, _ := json.Marshal(reqBody)

	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.StatusInProgress,
	}

	mockService.On("CreateReception", mock.Anything, pvzID).Return(reception, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/receptions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_AddProduct(t *testing.T) {
	mockService := new(MockService)
	cfg := &config.Config{}
	h := NewHandler(mockService, cfg)

	router := gin.Default()
	router.POST("/api/products", h.authMiddleware, h.addProduct)

	pvzID := uuid.New()
	reqBody := map[string]interface{}{
		"pvz_id":       pvzID.String(),
		"product_type": "electronics",
	}
	jsonBody, _ := json.Marshal(reqBody)

	product := &models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        models.TypeElectronics,
		ReceptionID: uuid.New(),
	}

	mockService.On("AddProduct", mock.Anything, pvzID, models.TypeElectronics).Return(product, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/products", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

// Тест с неверной ролью
func TestHandler_Register_InvalidRole(t *testing.T) {
	mockService := new(MockService)
	cfg := &config.Config{}
	h := NewHandler(mockService, cfg)

	router := gin.Default()
	router.POST("/auth/register", h.register)

	reqBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password",
		"role":     "invalid",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

// Тест с существующим email
func TestHandler_Register_ExistingEmail(t *testing.T) {
	mockService := new(MockService)
	cfg := &config.Config{}
	h := NewHandler(mockService, cfg)

	router := gin.Default()
	router.POST("/auth/register", h.register)

	mockService.On("RegisterUser", mock.Anything, "existing@example.com", "password", models.RoleEmployee).Return(nil, errors.New("user already exists"))

	reqBody := map[string]interface{}{
		"email":    "existing@example.com",
		"password": "password",
		"role":     "employee",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}
