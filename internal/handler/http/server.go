package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/avito/pvz/internal/service/product"
	"github.com/avito/pvz/internal/service/pvz"
	"github.com/avito/pvz/internal/service/reception"
	"github.com/avito/pvz/internal/service/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server представляет HTTP-сервер
type Server struct {
	server   *http.Server
	handlers *Handlers
}

// NewServer создает новый экземпляр Server
func NewServer(
	port int,
	pvzService *pvz.Service,
	receptionService *reception.Service,
	productService *product.Service,
	userService *user.Service,
) *Server {
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// Инициализация хендлеров
	handlers := NewHandlers(pvzService, receptionService, productService, userService)
	handlers.RegisterRoutes(router)

	return &Server{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: router,
		},
		handlers: handlers,
	}
}

// Start запускает HTTP-сервер
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Stop останавливает HTTP-сервер
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
