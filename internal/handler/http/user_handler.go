package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/avito/pvz/internal/domain/user"
	"github.com/avito/pvz/internal/handler/http/middleware"
	userService "github.com/avito/pvz/internal/service/user"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserHandler struct {
	service userService.ServiceInterface
}

func NewUserHandler(service userService.ServiceInterface) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// RegisterUser регистрирует нового пользователя
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userRole := user.Role(req.Role)
	newUser, err := h.service.Register(r.Context(), req.Email, req.Password, userRole)
	if err != nil {
		switch err {
		case userService.ErrInvalidEmail:
			http.Error(w, "email cannot be empty", http.StatusBadRequest)
		case userService.ErrInvalidPassword:
			http.Error(w, "password cannot be empty", http.StatusBadRequest)
		case userService.ErrInvalidRole:
			http.Error(w, "invalid role", http.StatusBadRequest)
		case userService.ErrUserAlreadyExists:
			http.Error(w, "user already exists", http.StatusConflict)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	response := struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Role      string `json:"role"`
		CreatedAt string `json:"created_at"`
	}{
		ID:        newUser.ID.String(),
		Email:     newUser.Email,
		Role:      string(newUser.Role),
		CreatedAt: newUser.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// LoginUser выполняет вход пользователя
func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "неверный формат запроса",
		})
		return
	}

	user, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case userService.ErrUserNotFound, userService.ErrInvalidPassword:
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "неверные учетные данные",
			})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "внутренняя ошибка сервера",
			})
		}
		return
	}

	response := map[string]interface{}{
		"email": user.Email,
		"role":  user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUser получает пользователя по ID
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	u, err := h.service.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Role      string `json:"role"`
		CreatedAt string `json:"created_at"`
	}{
		ID:        u.ID.String(),
		Email:     u.Email,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateUser обновляет данные пользователя
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" && req.Role == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	u := &user.User{
		ID:    userID,
		Email: req.Email,
		Role:  user.Role(req.Role),
	}

	if err := h.service.Update(r.Context(), u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteUser удаляет пользователя
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListUsers возвращает список пользователей
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if limit <= 0 {
		limit = 10
	}

	users, err := h.service.List(r.Context(), offset, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Role      string `json:"role"`
		CreatedAt string `json:"created_at"`
	}, len(users))

	for i, u := range users {
		response[i] = struct {
			ID        string `json:"id"`
			Email     string `json:"email"`
			Role      string `json:"role"`
			CreatedAt string `json:"created_at"`
		}{
			ID:        u.ID.String(),
			Email:     u.Email,
			Role:      string(u.Role),
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes регистрирует маршруты для пользователей
func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Use(middleware.RequireRole(user.RoleAdmin))

		r.Put("/user/{id}", h.UpdateUser)
		r.Delete("/user/{id}", h.DeleteUser)
		r.Get("/user", h.ListUsers)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Get("/user/{id}", h.GetUser)
	})

	// Публичные маршруты
	r.Post("/user/register", h.RegisterUser)
	r.Post("/user/login", h.LoginUser)
}
