package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/avito/pvz/internal/config"
	"github.com/avito/pvz/internal/domain/transaction"
	"github.com/avito/pvz/internal/domain/user"
	httphandler "github.com/avito/pvz/internal/handler/http"
	"github.com/avito/pvz/internal/middleware"
	"github.com/avito/pvz/internal/repository/postgres"
	"github.com/avito/pvz/internal/service/pvz"
	"github.com/avito/pvz/internal/service/reception"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// App представляет собой приложение
type App struct {
	server *http.Server
	router *mux.Router
}

// New создает новый экземпляр приложения
func New(cfg *config.Config) (*App, error) {
	// Инициализация репозиториев
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
	txManager := transaction.NewManager(sqlxDB)

	// Создаем реализацию аудита
	auditLog := postgres.NewAuditLog(sqlxDB)

	// Создаем модель пользователя по умолчанию
	defaultUser := &user.User{
		Role: user.RoleAdmin,
	}

	pvzService := pvz.New(pvzRepo, userRepo, txManager, auditLog, defaultUser)
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
