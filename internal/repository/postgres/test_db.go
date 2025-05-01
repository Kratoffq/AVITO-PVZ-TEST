package postgres

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// NewTestDB создает подключение к тестовой базе данных
func NewTestDB() (*sqlx.DB, error) {
	// Получаем параметры подключения из переменных окружения или используем значения по умолчанию
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5436")
	user := getEnv("TEST_DB_USER", "postgres")
	password := getEnv("TEST_DB_PASSWORD", "postgres")
	dbname := getEnv("TEST_DB_NAME", "avito_pvz_test")

	// Формируем строку подключения
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Подключаемся к базе данных
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping test database: %w", err)
	}

	// Инициализируем схему базы данных
	if err := initTestDB(db.DB); err != nil {
		return nil, fmt.Errorf("failed to initialize test database: %w", err)
	}

	return db, nil
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// initTestDB инициализирует схему тестовой базы данных
func initTestDB(db *sql.DB) error {
	_, err := db.Exec(`
		-- Создание таблицы пользователей
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) UNIQUE,
			password VARCHAR(255),
			role VARCHAR(50) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		);

		-- Создание таблицы ПВЗ
		CREATE TABLE IF NOT EXISTS pvzs (
			id UUID PRIMARY KEY,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			city VARCHAR(100) NOT NULL,
			CONSTRAINT city_check CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань'))
		);

		-- Создание таблицы приемок
		CREATE TABLE IF NOT EXISTS receptions (
			id UUID PRIMARY KEY,
			date_time TIMESTAMP WITH TIME ZONE NOT NULL,
			pvz_id UUID NOT NULL REFERENCES pvzs(id),
			status VARCHAR(50) NOT NULL DEFAULT 'in_progress',
			CONSTRAINT status_check CHECK (status IN ('in_progress', 'close'))
		);

		-- Создание таблицы товаров
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY,
			date_time TIMESTAMP WITH TIME ZONE NOT NULL,
			type VARCHAR(50) NOT NULL,
			reception_id UUID NOT NULL REFERENCES receptions(id),
			CONSTRAINT type_check CHECK (type IN ('electronics', 'clothing', 'food', 'other'))
		);

		-- Создание индексов
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_receptions_pvz_id ON receptions(pvz_id);
		CREATE INDEX IF NOT EXISTS idx_receptions_status ON receptions(status);
		CREATE INDEX IF NOT EXISTS idx_products_reception_id ON products(reception_id);
		CREATE INDEX IF NOT EXISTS idx_receptions_date_time ON receptions(date_time);
	`)
	return err
}

// SetupTestDB подготавливает тестовую базу данных
func SetupTestDB(t *testing.T) *sqlx.DB {
	db, err := NewTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Очищаем таблицы перед тестом
	_, err = db.Exec(`
		TRUNCATE TABLE products CASCADE;
		TRUNCATE TABLE receptions CASCADE;
		TRUNCATE TABLE pvzs CASCADE;
		TRUNCATE TABLE users CASCADE;
	`)
	if err != nil {
		t.Fatalf("Failed to clean test database: %v", err)
	}

	return db
}
