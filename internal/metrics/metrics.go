package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Технические метрики
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"method", "endpoint"},
	)

	// Бизнес метрики
	PvzCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pvz_created_total",
			Help: "Total number of created PVZs",
		},
	)

	ReceptionsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "receptions_created_total",
			Help: "Total number of created receptions",
		},
	)

	ProductsAddedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "products_added_total",
			Help: "Total number of added products",
		},
	)
)
