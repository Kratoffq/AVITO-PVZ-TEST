//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -package openapi -generate types,server,spec -o ../../api/openapi/types.gen.go ../../api/openapi/swagger.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -package http -generate types,server,spec -o ./openapi.gen.go ../../api/openapi/swagger.yaml

package http

import (
	"encoding/json"
	"net/http"

	domainUser "github.com/avito/pvz/internal/domain/user"
	"github.com/avito/pvz/internal/handler/http/middleware"
	serviceProduct "github.com/avito/pvz/internal/service/product"
	servicePVZ "github.com/avito/pvz/internal/service/pvz"
	serviceReception "github.com/avito/pvz/internal/service/reception"
	serviceUser "github.com/avito/pvz/internal/service/user"
	"github.com/avito/pvz/pkg/httpresponse"
	"github.com/go-chi/chi/v5"
)

// Handler содержит все HTTP обработчики
type Handler struct {
	pvzService       *servicePVZ.Service
	receptionService *serviceReception.Service
	productService   *serviceProduct.Service
	userService      *serviceUser.Service
}

// New создает новый экземпляр Handler
func New(pvzService *servicePVZ.Service, receptionService *serviceReception.Service, productService *serviceProduct.Service, userService *serviceUser.Service) *Handler {
	return &Handler{
		pvzService:       pvzService,
		receptionService: receptionService,
		productService:   productService,
		userService:      userService,
	}
}

// RegisterRoutes регистрирует маршруты
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Маршруты для регистрации и авторизации
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)

	// Маршруты, требующие авторизации
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		// Здесь будут регистрироваться маршруты других хендлеров
	})
}

// Register обрабатывает регистрацию пользователя
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string          `json:"email"`
		Password string          `json:"password"`
		Role     domainUser.Role `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	user, err := h.userService.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		switch err {
		case serviceUser.ErrInvalidEmail:
			httpresponse.Error(w, http.StatusBadRequest, "неверный формат email")
		case serviceUser.ErrInvalidPassword:
			httpresponse.Error(w, http.StatusBadRequest, "неверный формат пароля")
		case serviceUser.ErrInvalidRole:
			httpresponse.Error(w, http.StatusBadRequest, "неверная роль")
		case serviceUser.ErrUserAlreadyExists:
			httpresponse.Error(w, http.StatusConflict, "пользователь уже существует")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "ошибка при регистрации")
		}
		return
	}

	httpresponse.JSON(w, http.StatusCreated, user)
}

// Login обрабатывает авторизацию пользователя
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	token, err := h.userService.LoginUser(r.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case serviceUser.ErrUserNotFound:
			httpresponse.Error(w, http.StatusUnauthorized, "неверный email или пароль")
		case serviceUser.ErrInvalidPassword:
			httpresponse.Error(w, http.StatusUnauthorized, "неверный email или пароль")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "ошибка при авторизации")
		}
		return
	}

	httpresponse.JSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}

// Handlers содержит все HTTP-хендлеры
type Handlers struct {
	PVZ       *PVZHandler
	Reception *ReceptionHandler
	Product   *ProductHandler
	User      *UserHandler
}

// NewHandlers создает новый экземпляр Handlers
func NewHandlers(
	pvzService *servicePVZ.Service,
	receptionService *serviceReception.Service,
	productService *serviceProduct.Service,
	userService *serviceUser.Service,
) *Handlers {
	return &Handlers{
		PVZ:       NewPVZHandler(pvzService),
		Reception: NewReceptionHandler(receptionService),
		Product:   NewProductHandler(productService),
		User:      NewUserHandler(userService),
	}
}

// RegisterRoutes регистрирует все маршруты
func (h *Handlers) RegisterRoutes(r chi.Router) {
	h.PVZ.RegisterRoutes(r)
	h.Reception.RegisterRoutes(r)
	h.Product.RegisterRoutes(r)
	h.User.RegisterRoutes(r)
}
