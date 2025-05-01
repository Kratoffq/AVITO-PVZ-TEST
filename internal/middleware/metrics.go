package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/avito/pvz/internal/metrics"
	"github.com/gorilla/mux"
)

// MetricsMiddleware добавляет сбор метрик для HTTP запросов
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем ResponseWriter для перехвата статуса ответа
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Выполняем следующий обработчик
		next.ServeHTTP(rw, r)

		// Получаем путь из маршрута
		var path string
		if route := mux.CurrentRoute(r); route != nil {
			path, _ = route.GetPathTemplate()
		} else {
			path = r.URL.Path
		}

		// Обновляем метрики
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rw.statusCode)

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	})
}

// responseWriter перехватывает статус ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader перехватывает статус ответа
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
