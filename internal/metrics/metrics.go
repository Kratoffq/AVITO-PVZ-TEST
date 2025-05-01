package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP метрики
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Общее количество HTTP запросов",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Длительность HTTP запросов в секундах",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Бизнес метрики
	PVZCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pvz_created_total",
			Help: "Общее количество созданных ПВЗ",
		},
	)

	ReceptionCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "reception_created_total",
			Help: "Общее количество созданных приёмок",
		},
	)

	ProductCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "product_created_total",
			Help: "Общее количество созданных товаров",
		},
	)

	// Метрики транзакций
	TransactionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "transaction_duration_seconds",
			Help:    "Длительность транзакций в секундах",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)

	TransactionErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_errors_total",
			Help: "Общее количество ошибок в транзакциях",
		},
		[]string{"type"},
	)
)
