package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config содержит конфигурацию приложения
type Config struct {
	HTTP struct {
		Port int
	}
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
		SSLMode  string
	}
}

// DBConfig содержит конфигурацию подключения к базе данных
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	cfg := &Config{}

	// HTTP конфигурация
	port, err := strconv.Atoi(getEnv("HTTP_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_PORT: %w", err)
	}
	cfg.HTTP.Port = port

	// Database конфигурация
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	port, err = strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}
	cfg.Database.Port = port
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "postgres")
	cfg.Database.DBName = getEnv("DB_NAME", "pvz")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")

	return cfg, nil
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
