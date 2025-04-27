package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Вызов следующего обработчика
		c.Next()

		// Сбор метрик после обработки запроса
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			strconv.Itoa(status),
		).Inc()

		HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}
