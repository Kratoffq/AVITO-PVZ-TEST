package monitoring

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware - middleware для логирования HTTP-запросов
type LoggingMiddleware struct {
	logger *log.Logger
}

// NewLoggingMiddleware создает новый экземпляр LoggingMiddleware
func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Middleware возвращает HTTP middleware для логирования запросов
func (m *LoggingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем ResponseWriter для отслеживания статуса ответа
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		// Логируем запрос
		duration := time.Since(start)
		m.logger.Printf(
			"[%s] %s %s %d %s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			rw.statusCode,
			duration,
		)
	})
}

// MetricsMiddleware - middleware для сбора метрик HTTP-запросов
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем ResponseWriter для отслеживания статуса ответа
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		// Собираем метрики
		duration := time.Since(start).Seconds()
		ObserveHTTPRequestDuration(r.Method, r.URL.Path, duration)
		IncHTTPRequest(r.Method, r.URL.Path, http.StatusText(rw.statusCode))
	})
}

// responseWriter - обертка над http.ResponseWriter для отслеживания статуса ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
