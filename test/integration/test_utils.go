package integration

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

const (
	testDBHost     = "localhost"
	testDBPort     = 5434
	testDBUser     = "postgres"
	testDBPassword = "postgres"
	testDBName     = "pvz_test"
)

func setupTestDB() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPassword, testDBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping test database: %w", err)
	}

	// Применение миграций
	if err := applyMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Очистка таблиц перед тестом
	tables := []string{"products", "receptions", "pvzs", "users"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			return nil, fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	return db, nil
}

func applyMigrations(db *sql.DB) error {
	// Получаем текущую директорию
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Путь к файлу миграции относительно корня проекта
	migrationPath := filepath.Join(wd, "..", "..", "internal", "repository", "migration", "000001_init.up.sql")
	migrationSQL, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Применение миграции
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		return fmt.Errorf("failed to apply migration: %w", err)
	}

	return nil
}
