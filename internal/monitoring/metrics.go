package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP метрики
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// Метрики базы данных
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	dbErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation"},
	)

	// Метрики для операций с ПВЗ
	pvzOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pvz_operations_total",
			Help: "Total number of PVZ operations",
		},
		[]string{"operation", "status"},
	)

	pvzOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pvz_operation_duration_seconds",
			Help:    "Duration of PVZ operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Метрики для операций с товарами
	productOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_operations_total",
			Help: "Total number of product operations",
		},
		[]string{"operation", "status"},
	)

	productOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "product_operation_duration_seconds",
			Help:    "Duration of product operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Метрики для операций с приемками
	receptionOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "reception_operations_total",
			Help: "Total number of reception operations",
		},
		[]string{"operation", "status"},
	)

	receptionOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "reception_operation_duration_seconds",
			Help:    "Duration of reception operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Добавляем новые метрики
	pvzCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pvz_created_total",
		Help: "Total number of created PVZ",
	})

	receptionsCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "receptions_created_total",
		Help: "Total number of created receptions",
	})

	productsAddedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_added_total",
		Help: "Total number of added products",
	})
)

// IncHTTPRequest увеличивает счетчик HTTP запросов
func IncHTTPRequest(method, path, status string) {
	httpRequestsTotal.WithLabelValues(method, path, status).Inc()
}

// ObserveHTTPRequestDuration записывает длительность HTTP запроса
func ObserveHTTPRequestDuration(method, path string, duration float64) {
	httpRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// IncPVZOperation увеличивает счетчик операций с ПВЗ
func IncPVZOperation(operation, status string) {
	pvzOperationsTotal.WithLabelValues(operation, status).Inc()
}

// ObservePVZOperationDuration записывает длительность операции с ПВЗ
func ObservePVZOperationDuration(operation string, duration float64) {
	pvzOperationDuration.WithLabelValues(operation).Observe(duration)
}

// IncProductOperation увеличивает счетчик операций с товарами
func IncProductOperation(operation, status string) {
	productOperationsTotal.WithLabelValues(operation, status).Inc()
}

// ObserveProductOperationDuration записывает длительность операции с товаром
func ObserveProductOperationDuration(operation string, duration float64) {
	productOperationDuration.WithLabelValues(operation).Observe(duration)
}

// IncReceptionOperation увеличивает счетчик операций с приемками
func IncReceptionOperation(operation, status string) {
	receptionOperationsTotal.WithLabelValues(operation, status).Inc()
}

// ObserveReceptionOperationDuration записывает длительность операции с приемкой
func ObserveReceptionOperationDuration(operation string, duration float64) {
	receptionOperationDuration.WithLabelValues(operation).Observe(duration)
}

// ObserveDBQueryDuration записывает длительность запроса к базе данных
func ObserveDBQueryDuration(operation string, duration float64) {
	dbQueryDuration.WithLabelValues(operation).Observe(duration)
}

// IncDBError увеличивает счетчик ошибок базы данных
func IncDBError(operation string) {
	dbErrorsTotal.WithLabelValues(operation).Inc()
}

// IncPVZCreated увеличивает счетчик созданных ПВЗ
func IncPVZCreated() {
	pvzCreatedTotal.Inc()
}

// IncReceptionCreated увеличивает счетчик созданных приемок
func IncReceptionCreated() {
	receptionsCreatedTotal.Inc()
}

// IncProductAdded увеличивает счетчик добавленных товаров
func IncProductAdded() {
	productsAddedTotal.Inc()
}
