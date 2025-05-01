package app

import (
	"fmt"

	"github.com/avito/pvz/api/proto"
	"github.com/avito/pvz/internal/domain/transaction"
	"github.com/avito/pvz/internal/domain/user"
	"github.com/avito/pvz/internal/handler/grpc"
	"github.com/avito/pvz/internal/repository/postgres"
	"github.com/avito/pvz/internal/service/pvz"
	"github.com/jmoiron/sqlx"
	grpcserver "google.golang.org/grpc"
)

// NewGRPCServer создает новый экземпляр gRPC сервера
func NewGRPCServer(cfg *Config) (*grpcserver.Server, error) {
	// Инициализация репозиториев
	db, err := postgres.New(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// Конвертируем sql.DB в sqlx.DB
	sqlxDB := sqlx.NewDb(db.DB, "postgres")

	// Инициализация сервисов
	pvzRepo := postgres.NewPVZRepository(sqlxDB)
	userRepo := postgres.NewUserRepository(sqlxDB)
	txManager := transaction.NewManager(sqlxDB)
	auditLog := postgres.NewAuditLog(sqlxDB)

	// Создаем модель пользователя по умолчанию
	defaultUser := &user.User{
		Role: user.RoleAdmin,
	}

	pvzService := pvz.New(pvzRepo, userRepo, txManager, auditLog, defaultUser)

	// Создание gRPC сервера
	server := grpcserver.NewServer()

	// Регистрация сервисов
	pvzHandler := grpc.NewPVZHandler(pvzService)
	proto.RegisterPVZServiceServer(server, pvzHandler)

	return server, nil
}
