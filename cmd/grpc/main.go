package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/avito/pvz/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Создаем конфигурацию
	cfg := &app.Config{
		Server: struct {
			HTTP struct {
				Host string
				Port int
			}
			GRPC struct {
				Host string
				Port int
			}
		}{
			GRPC: struct {
				Host string
				Port int
			}{
				Host: "localhost",
				Port: 3000,
			},
		},
	}

	// Создаем gRPC сервер
	server, err := app.NewGRPCServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}

	// Запускаем сервер в горутине
	go func() {
		listener, err := net.Listen("tcp", ":3000")
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	log.Printf("gRPC server is running on port %d", cfg.Server.GRPC.Port)

	// Ожидаем сигнал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	server.GracefulStop()
	log.Println("gRPC server stopped gracefully")
}
