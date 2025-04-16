package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/avito/pvz/internal/config"
	"github.com/avito/pvz/internal/handler"
	"github.com/avito/pvz/internal/repository/postgres"
	"github.com/avito/pvz/internal/service/impl"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %s", err.Error())
	}
	logrus.Info("Config loaded successfully")

	db, err := postgres.NewPostgresDB(cfg.DBConfig)
	if err != nil {
		logrus.Fatalf("failed to initialize db: %s", err.Error())
	}
	defer db.Close()

	// Настройка пула соединений
	db.SetMaxOpenConns(cfg.DBConfig.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DBConfig.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConfig.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.DBConfig.ConnMaxIdleTime)

	logrus.Info("Database connection pool configured")

	repos := postgres.NewRepository(db)
	logrus.Info("Repository initialized")

	services := impl.NewService(repos, cfg)
	logrus.Info("Service layer initialized")

	handlers := handler.NewHandler(services, cfg)
	logrus.Info("Handlers initialized")

	// Инициализация роутера
	router := gin.New()
	router = handlers.InitRoutes()
	logrus.Info("Routes initialized")

	// Настройка HTTP сервера
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerConfig.Port),
		Handler:      router,
		ReadTimeout:  cfg.ServerConfig.ReadTimeout,
		WriteTimeout: cfg.ServerConfig.WriteTimeout,
		IdleTimeout:  cfg.ServerConfig.IdleTimeout,
	}

	// Запуск сервера метрик Prometheus
	metricsSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Prometheus.Port),
		Handler: promhttp.Handler(),
	}

	// Запуск серверов в горутинах
	go func() {
		logrus.Infof("Starting HTTP server on port %d", cfg.ServerConfig.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("error occurred while running http server: %s", err.Error())
		}
	}()

	go func() {
		logrus.Infof("Starting metrics server on port %d", cfg.Prometheus.Port)
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Errorf("error occurred while running metrics server: %s", err.Error())
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Info("Server Shutting Down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("Server forced to shutdown: %s", err)
	}

	if err := metricsSrv.Shutdown(ctx); err != nil {
		logrus.Errorf("Metrics server forced to shutdown: %s", err)
	}

	logrus.Info("Server exited properly")
}
