package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/avito/pvz/internal/config"
	"github.com/avito/pvz/internal/domain/audit"
	"github.com/avito/pvz/internal/domain/pvz"
	"github.com/avito/pvz/internal/domain/transaction"
	"github.com/avito/pvz/internal/domain/user"
	httphandler "github.com/avito/pvz/internal/handler/http"
	"github.com/avito/pvz/internal/middleware"
	"github.com/avito/pvz/internal/repository/postgres"
	servicePVZ "github.com/avito/pvz/internal/service/pvz"
	"github.com/avito/pvz/internal/service/reception"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// App представляет собой приложение
type App struct {
	server *http.Server
	router *mux.Router
	config *config.Config
}

// New создает новый экземпляр приложения
func New(cfg *config.Config) (*App, error) {
	// Инициализация базы данных
	db, err := postgres.New(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// Конвертируем sql.DB в sqlx.DB
	sqlxDB := sqlx.NewDb(db.DB, "postgres")

	// Инициализация репозиториев
	pvzRepo := postgres.NewPVZRepository(sqlxDB)
	userRepo := postgres.NewUserRepository(sqlxDB)
	receptionRepo := postgres.NewReceptionRepository(sqlxDB)
	productRepo := postgres.NewProductRepository(sqlxDB)
	auditLog := postgres.NewAuditLog(sqlxDB)

	// Инициализация менеджера транзакций
	txManager := transaction.NewManager(sqlxDB)

	// Создание сервисов
	pvzService := servicePVZ.New(pvzRepo, userRepo, txManager, auditLog, nil)
	receptionService := reception.New(receptionRepo, pvzRepo, txManager, productRepo)

	// Создаем роутер
	router := mux.NewRouter()

	// Добавляем middleware
	router.Use(middleware.MetricsMiddleware)

	// Регистрируем обработчики
	httphandler.RegisterPVZHandlers(router, pvzService)
	httphandler.RegisterReceptionHandlers(router, receptionService)

	// Создаем HTTP сервер
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler: router,
	}

	return &App{
		server: server,
		router: router,
		config: cfg,
	}, nil
}

// Router возвращает HTTP роутер
func (a *App) Router() *mux.Router {
	return a.router
}

// Start запускает приложение
func (a *App) Start() error {
	return a.server.ListenAndServe()
}

// Stop останавливает приложение
func (a *App) Stop(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

// NewPVZService создает новый экземпляр сервиса PVZ
func (a *App) NewPVZService(pvzRepo pvz.Repository, userRepo user.Repository, txManager transaction.Manager, auditLog audit.AuditLog) *servicePVZ.Service {
	return servicePVZ.New(pvzRepo, userRepo, txManager, auditLog, nil)
}
