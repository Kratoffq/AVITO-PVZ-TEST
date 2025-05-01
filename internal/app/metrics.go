package app

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartMetricsServer запускает сервер метрик на порту 9000
func StartMetricsServer() error {
	http.Handle("/metrics", promhttp.Handler())

	addr := ":9000"
	fmt.Printf("Starting metrics server on %s\n", addr)

	return http.ListenAndServe(addr, nil)
}
