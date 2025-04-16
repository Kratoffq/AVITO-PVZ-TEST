package monitoring

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsMiddleware(t *testing.T) {
	// Создаем тестовый хендлер
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// Применяем middleware
	middleware := MetricsMiddleware(handler)
	middleware.ServeHTTP(rec, req)

	// Проверяем, что запрос обработан
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	// Создаем тестовый логгер
	var logOutput bytes.Buffer
	logger := log.New(&logOutput, "", 0)

	// Создаем тестовый хендлер
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	rec := httptest.NewRecorder()

	// Применяем middleware
	middleware := NewLoggingMiddleware(logger).Middleware(handler)
	middleware.ServeHTTP(rec, req)

	// Проверяем, что запрос обработан
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	// Проверяем, что запрос залогирован
	logStr := logOutput.String()
	expectedParts := []string{
		"[GET]",
		"/test",
		"192.0.2.1:1234",
		"200",
	}
	for _, part := range expectedParts {
		if !strings.Contains(logStr, part) {
			t.Errorf("Expected log to contain '%s', got: %s", part, logStr)
		}
	}
}
