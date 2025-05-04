package app

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/avito/pvz/internal/config"
	"github.com/avito/pvz/internal/domain/pvz"
	"github.com/avito/pvz/internal/domain/user"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDB представляет мок базы данных
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Begin() (*sql.Tx, error) {
	args := m.Called()
	return args.Get(0).(*sql.Tx), args.Error(1)
}

func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDB) Ping() error {
	args := m.Called()
	return args.Error(0)
}

// MockHTTPServer представляет мок HTTP сервера
type MockHTTPServer struct {
	mock.Mock
}

func (m *MockHTTPServer) ListenAndServe() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHTTPServer) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockPVZRepository представляет мок репозитория PVZ
type MockPVZRepository struct {
	mock.Mock
}

func (m *MockPVZRepository) GetAll(ctx context.Context) ([]*pvz.PVZ, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) GetByID(ctx context.Context, id uuid.UUID) (*pvz.PVZ, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) Create(ctx context.Context, pvz *pvz.PVZ) error {
	args := m.Called(ctx, pvz)
	return args.Error(0)
}

func (m *MockPVZRepository) Update(ctx context.Context, pvz *pvz.PVZ) error {
	args := m.Called(ctx, pvz)
	return args.Error(0)
}

func (m *MockPVZRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPVZRepository) GetWithReceptions(ctx context.Context, from, to time.Time, offset, limit int) ([]*pvz.PVZWithReceptions, error) {
	args := m.Called(ctx, from, to, offset, limit)
	return args.Get(0).([]*pvz.PVZWithReceptions), args.Error(1)
}

func (m *MockPVZRepository) GetByCity(ctx context.Context, city string) (*pvz.PVZ, error) {
	args := m.Called(ctx, city)
	return args.Get(0).(*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) List(ctx context.Context, offset, limit int) ([]*pvz.PVZ, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*pvz.PVZ), args.Error(1)
}

// MockUserRepository представляет мок репозитория пользователей
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*user.User, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// MockTransactionManager представляет мок менеджера транзакций
type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

// MockAuditLog представляет мок аудита
type MockAuditLog struct {
	mock.Mock
}

func (m *MockAuditLog) LogPVZCreation(ctx context.Context, pvzID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, pvzID, userID)
	return args.Error(0)
}

func (m *MockAuditLog) LogPVZUpdate(ctx context.Context, pvzID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, pvzID, userID)
	return args.Error(0)
}

func (m *MockAuditLog) LogPVZDeletion(ctx context.Context, pvzID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, pvzID, userID)
	return args.Error(0)
}

func TestNew(t *testing.T) {
	// Создаем мок базы данных
	mockDB := &MockDB{}
	mockDB.On("Ping").Return(nil)

	// Создаем конфигурацию
	cfg := &config.Config{
		HTTP: struct{ Port int }{Port: 8080},
		Database: struct {
			Host     string
			Port     int
			User     string
			Password string
			DBName   string
			SSLMode  string
		}{
			Host:     "localhost",
			Port:     5434,
			User:     "postgres",
			Password: "postgres",
			DBName:   "pvz_test",
			SSLMode:  "disable",
		},
	}

	// Создаем приложение
	app, err := New(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, app.router)
	assert.NotNil(t, app.server)
	assert.Equal(t, cfg, app.config)
}

func TestRouter(t *testing.T) {
	app := &App{
		router: nil,
	}
	assert.Nil(t, app.Router())

	app.router = &mux.Router{}
	assert.NotNil(t, app.Router())
}

func TestStart(t *testing.T) {
	mockServer := new(MockHTTPServer)
	app := &App{
		server: &http.Server{},
	}

	// Тест успешного запуска
	mockServer.On("ListenAndServe").Return(nil)
	err := app.Start()
	assert.NoError(t, err)

	// Тест ошибки при запуске
	mockServer.On("ListenAndServe").Return(http.ErrServerClosed)
	err = app.Start()
	assert.Error(t, err)
}

func TestStop(t *testing.T) {
	mockServer := new(MockHTTPServer)
	app := &App{
		server: &http.Server{},
	}

	ctx := context.Background()

	// Тест успешной остановки
	mockServer.On("Shutdown", ctx).Return(nil)
	err := app.Stop(ctx)
	assert.NoError(t, err)

	// Тест остановки с отмененным контекстом
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = app.Stop(ctx)
	assert.Error(t, err)
}

func TestNewPVZService(t *testing.T) {
	app := &App{}

	// Создаем моки для зависимостей
	mockPVZRepo := new(MockPVZRepository)
	mockUserRepo := new(MockUserRepository)
	mockTxManager := new(MockTransactionManager)
	mockAuditLog := new(MockAuditLog)

	service := app.NewPVZService(mockPVZRepo, mockUserRepo, mockTxManager, mockAuditLog)
	assert.NotNil(t, service)
}

func TestApp_StartStop(t *testing.T) {
	cfg := &config.Config{
		HTTP: struct{ Port int }{
			Port: 8081,
		},
		Database: struct {
			Host     string
			Port     int
			User     string
			Password string
			DBName   string
			SSLMode  string
		}{
			Host:     "localhost",
			Port:     5434,
			User:     "postgres",
			Password: "postgres",
			DBName:   "pvz_test",
			SSLMode:  "disable",
		},
	}

	app, err := New(cfg)
	require.NoError(t, err)

	// Запускаем приложение в горутине
	go func() {
		err := app.Start()
		if err != nil && err.Error() != "http: Server closed" {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	// Даем приложению время на запуск
	time.Sleep(100 * time.Millisecond)

	// Останавливаем приложение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = app.Stop(ctx)
	assert.NoError(t, err)
}

func TestApp_Router(t *testing.T) {
	cfg := &config.Config{
		HTTP: struct{ Port int }{
			Port: 8082,
		},
		Database: struct {
			Host     string
			Port     int
			User     string
			Password string
			DBName   string
			SSLMode  string
		}{
			Host:     "localhost",
			Port:     5434,
			User:     "postgres",
			Password: "postgres",
			DBName:   "pvz_test",
			SSLMode:  "disable",
		},
	}

	app, err := New(cfg)
	require.NoError(t, err)

	router := app.Router()
	assert.NotNil(t, router)
	assert.Equal(t, app.router, router)
}

func TestApp_NewPVZService(t *testing.T) {
	cfg := &config.Config{
		HTTP: struct{ Port int }{
			Port: 8083,
		},
		Database: struct {
			Host     string
			Port     int
			User     string
			Password string
			DBName   string
			SSLMode  string
		}{
			Host:     "localhost",
			Port:     5434,
			User:     "postgres",
			Password: "postgres",
			DBName:   "pvz_test",
			SSLMode:  "disable",
		},
	}

	app, err := New(cfg)
	require.NoError(t, err)

	// Создаем моки для зависимостей
	pvzRepo := &MockPVZRepository{}
	userRepo := &MockUserRepository{}
	txManager := &MockTransactionManager{}
	auditLog := &MockAuditLog{}

	// Создаем сервис
	service := app.NewPVZService(pvzRepo, userRepo, txManager, auditLog)
	assert.NotNil(t, service)
}
