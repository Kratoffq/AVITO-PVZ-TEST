package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/avito/pvz/internal/app"
	"github.com/avito/pvz/internal/config"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем приложение
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Запускаем сервер метрик в отдельной горутине
	go func() {
		if err := app.StartMetricsServer(); err != nil {
			log.Printf("Failed to start metrics server: %v", err)
		}
	}()

	// Запускаем HTTP сервер
	go func() {
		if err := application.Start(); err != nil {
			log.Printf("Failed to start HTTP server: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Останавливаем приложение
	if err := application.Stop(context.Background()); err != nil {
		log.Printf("Failed to stop application: %v", err)
	}
}
