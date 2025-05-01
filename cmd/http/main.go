package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/avito/pvz/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Создаем конфигурацию
	cfg := &app.Config{
		Server: struct {
			HTTP struct {
				Host string
				Port int
			}
			GRPC struct {
				Host string
				Port int
			}
		}{
			HTTP: struct {
				Host string
				Port int
			}{
				Host: getEnv("HTTP_HOST", "localhost"),
				Port: getEnvAsInt("HTTP_PORT", 8080),
			},
			GRPC: struct {
				Host string
				Port int
			}{
				Host: getEnv("GRPC_HOST", "localhost"),
				Port: getEnvAsInt("GRPC_PORT", 9090),
			},
		},
		Database: struct {
			Host     string
			Port     int
			User     string
			Password string
			DBName   string
			SSLMode  string
		}{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "avito_pvz"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: struct {
			Secret     string
			Expiration string
		}{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			Expiration: getEnv("JWT_EXPIRATION", "24h"),
		},
		Logging: struct {
			Level  string
			Format string
		}{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	// Создаем HTTP-сервер
	server, err := app.NewHTTPServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create HTTP server: %v", err)
	}

	// Запускаем сервер в горутине
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Ожидаем сигнал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		log.Fatalf("Failed to stop HTTP server: %v", err)
	}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает значение переменной окружения как int или возвращает значение по умолчанию
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
