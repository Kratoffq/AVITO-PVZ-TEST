package app

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/avito/pvz/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig для тестов
type TestConfig struct {
	Server struct {
		HTTP struct {
			Host string
			Port int
		}
		GRPC struct {
			Host string
			Port int
		}
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

type MockConfig struct {
	Server struct {
		HTTP struct {
			Host string
			Port int
		}
	}
	Database *config.DBConfig
}

func TestNewHTTPServer(t *testing.T) {
	cfg := &Config{
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
				Host: "localhost",
				Port: 8080,
			},
			GRPC: struct {
				Host string
				Port int
			}{
				Host: "localhost",
				Port: 9090,
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
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "pvz_test",
			SSLMode:  "disable",
		},
		JWT: struct {
			Secret     string
			Expiration string
		}{
			Secret:     "test-secret",
			Expiration: "24h",
		},
		Logging: struct {
			Level  string
			Format string
		}{
			Level:  "debug",
			Format: "json",
		},
	}

	server, err := NewHTTPServer(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, server)
	assert.NotNil(t, server.server)
	assert.NotNil(t, server.router)
}

func TestHTTPServer_StartStop(t *testing.T) {
	cfg := &Config{
		Database: struct {
			Host     string
			Port     int
			User     string
			Password string
			DBName   string
			SSLMode  string
		}{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "pvz_test",
			SSLMode:  "disable",
		},
	}

	server, err := NewHTTPServer(cfg)
	require.NoError(t, err)

	// Запускаем сервер в горутине
	go func() {
		err := server.Start()
		if err != nil && err.Error() != "http: Server closed" {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	// Даем серверу время на запуск
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что сервер отвечает
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get("http://localhost:8082/health")
	if err == nil {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	// Останавливаем сервер
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Stop(ctx)
	assert.NoError(t, err)
}

func TestHTTPServerStart(t *testing.T) {
	server := &HTTPServer{
		server: &http.Server{
			Addr: ":8080",
		},
	}

	// Тест запуска сервера в горутине
	go func() {
		err := server.Start()
		assert.Error(t, err) // Ожидаем ошибку, так как сервер будет остановлен
	}()

	// Даем серверу время на запуск
	time.Sleep(100 * time.Millisecond)

	// Останавливаем сервер
	err := server.Stop(context.Background())
	assert.NoError(t, err)
}

func TestHTTPServerStop(t *testing.T) {
	server := &HTTPServer{
		server: &http.Server{
			Addr: ":8080",
		},
	}

	// Тест остановки сервера
	err := server.Stop(context.Background())
	assert.NoError(t, err)

	// Тест остановки с отмененным контекстом
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = server.Stop(ctx)
	assert.Error(t, err)
}
