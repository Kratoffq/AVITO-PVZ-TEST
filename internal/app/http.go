package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	domainuser "github.com/avito/pvz/internal/domain/user"
	httphandler "github.com/avito/pvz/internal/handler/http"
	"github.com/avito/pvz/internal/repository/postgres"
	"github.com/avito/pvz/internal/service/product"
	"github.com/avito/pvz/internal/service/pvz"
	"github.com/avito/pvz/internal/service/reception"
	userservice "github.com/avito/pvz/internal/service/user"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// HTTPServer представляет HTTP-сервер приложения
type HTTPServer struct {
	server *http.Server
	router chi.Router
}

// NewHTTPServer создает новый экземпляр HTTP-сервера
func NewHTTPServer(cfg *Config) (*HTTPServer, error) {
	// Инициализация репозиториев
	db, err := postgres.New(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// Конвертируем *sql.DB в *sqlx.DB
	sqlxDB := sqlx.NewDb(db.DB, "postgres")

	// Инициализация репозиториев
	pvzRepo := postgres.NewPVZRepository(sqlxDB)
	receptionRepo := postgres.NewReceptionRepository(sqlxDB)
	productRepo := postgres.NewProductRepository(sqlxDB)
	userRepo := postgres.NewUserRepository(sqlxDB)

	// Инициализация менеджера транзакций
	txManager := postgres.NewTransactionManager(db.DB)

	// Создаем реализацию аудита
	auditLog := postgres.NewAuditLog(sqlxDB)

	// Создаем модель пользователя по умолчанию
	defaultUser := &domainuser.User{
		Role: domainuser.RoleAdmin,
	}

	// Инициализация сервисов
	pvzService := pvz.New(pvzRepo, userRepo, txManager, auditLog, defaultUser)
	receptionService := reception.New(receptionRepo, pvzRepo, txManager, productRepo)
	productService := product.New(productRepo, receptionRepo, txManager)
	userService := userservice.New(userRepo, txManager)

	// Инициализация обработчиков
	handler := httphandler.New(pvzService, receptionService, productService, userService)

	// Настройка маршрутизатора
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	// Создание HTTP-сервера
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.HTTP.Host, cfg.Server.HTTP.Port),
		Handler:      router,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	return &HTTPServer{
		server: server,
		router: router,
	}, nil
}

// Start запускает HTTP-сервер
func (s *HTTPServer) Start() error {
	return s.server.ListenAndServe()
}

// Stop останавливает HTTP-сервер
func (s *HTTPServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
