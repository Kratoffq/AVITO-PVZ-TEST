package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/avito/pvz/internal/config"
	"github.com/avito/pvz/internal/handler"
	"github.com/avito/pvz/internal/repository/postgres"
	"github.com/avito/pvz/internal/service/impl"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// migrateDB выполняет миграцию схемы базы данных
func migrateDB(db *sql.DB) error {
	// Читаем SQL файл с миграцией
	migrationSQL, err := os.ReadFile("internal/repository/postgres/init.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %v", err)
	}

	// Выполняем миграцию
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %v", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация базы данных
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBConfig.Host,
		cfg.DBConfig.Port,
		cfg.DBConfig.User,
		cfg.DBConfig.Password,
		cfg.DBConfig.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Выполняем миграцию базы данных
	if err := migrateDB(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(cfg.DBConfig.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DBConfig.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConfig.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.DBConfig.ConnMaxIdleTime)

	// Инициализация репозитория
	repo := postgres.NewRepository(db)

	// Инициализация сервиса
	service := impl.NewService(repo, cfg)

	// Инициализация обработчика
	h := handler.NewHandler(service, cfg)
	router := h.InitRoutes()

	// Настройка сервера
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerConfig.Port),
		Handler:      router,
		ReadTimeout:  cfg.ServerConfig.ReadTimeout,
		WriteTimeout: cfg.ServerConfig.WriteTimeout,
		IdleTimeout:  cfg.ServerConfig.IdleTimeout,
	}

	// Запуск сервера
	log.Printf("Server starting on port %d", cfg.ServerConfig.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func registerRoutes(router *gin.Engine) {
	// TODO: Добавить маршруты
}
