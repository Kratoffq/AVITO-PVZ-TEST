package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	domainPVZ "github.com/avito/pvz/internal/domain/pvz"
	"github.com/avito/pvz/internal/handler/http/middleware"
	servicePVZ "github.com/avito/pvz/internal/service/pvz"
	"github.com/avito/pvz/pkg/auth"
	"github.com/avito/pvz/pkg/httpresponse"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// PVZServiceInterface определяет интерфейс для сервиса ПВЗ
type PVZServiceInterface interface {
	Create(ctx context.Context, city string, userID uuid.UUID) (*domainPVZ.PVZ, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domainPVZ.PVZ, error)
	GetWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*domainPVZ.PVZWithReceptions, error)
	GetAll(ctx context.Context) ([]*domainPVZ.PVZ, error)
	Update(ctx context.Context, pvz *domainPVZ.PVZ, moderatorID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID, moderatorID uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*domainPVZ.PVZ, error)
}

// PVZHandler обрабатывает HTTP-запросы для ПВЗ
type PVZHandler struct {
	service PVZServiceInterface
}

// NewPVZHandler создает новый экземпляр PVZHandler
func NewPVZHandler(service PVZServiceInterface) *PVZHandler {
	return &PVZHandler{
		service: service,
	}
}

// RegisterRoutes регистрирует маршруты для ПВЗ
func (h *PVZHandler) RegisterRoutes(r chi.Router) {
	r.Post("/pvz", h.Create)
	r.Get("/pvz/{id}", h.GetByID)
	r.Get("/pvz", h.GetWithReceptions)
}

// Create обрабатывает создание ПВЗ
func (h *PVZHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		City string `json:"city"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if req.City == "" {
		httpresponse.Error(w, http.StatusBadRequest, "город не может быть пустым")
		return
	}

	// Получаем ID пользователя из контекста
	userID, ok := auth.GetUserID(r.Context())
	if !ok || userID == uuid.Nil {
		httpresponse.Error(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}

	pvz, err := h.service.Create(r.Context(), req.City, userID)
	if err != nil {
		switch err {
		case servicePVZ.ErrInvalidCity:
			httpresponse.Error(w, http.StatusBadRequest, "неверное название города")
		case servicePVZ.ErrPVZAlreadyExists:
			httpresponse.Error(w, http.StatusConflict, "ПВЗ уже существует")
		case servicePVZ.ErrAccessDenied:
			httpresponse.Error(w, http.StatusForbidden, "доступ запрещен")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "ошибка при создании ПВЗ")
		}
		return
	}

	response := struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		City      string    `json:"city"`
	}{
		ID:        pvz.ID.String(),
		CreatedAt: pvz.CreatedAt,
		City:      pvz.City,
	}

	httpresponse.JSON(w, http.StatusCreated, response)
}

// GetByID обрабатывает получение ПВЗ по ID
func (h *PVZHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат ID ПВЗ")
		return
	}

	pvz, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		httpresponse.Error(w, http.StatusNotFound, "ПВЗ не найден")
		return
	}

	httpresponse.JSON(w, http.StatusOK, pvz)
}

// GetWithReceptions обрабатывает получение списка ПВЗ с приемками
func (h *PVZHandler) GetWithReceptions(w http.ResponseWriter, r *http.Request) {
	startDate, err := time.Parse(time.RFC3339, r.URL.Query().Get("start_date"))
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат даты начала")
		return
	}

	endDate, err := time.Parse(time.RFC3339, r.URL.Query().Get("end_date"))
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат даты окончания")
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if _, err := fmt.Sscanf(p, "%d", &page); err != nil {
			httpresponse.Error(w, http.StatusBadRequest, "неверный формат страницы")
			return
		}
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			httpresponse.Error(w, http.StatusBadRequest, "неверный формат лимита")
			return
		}
	}

	pvzs, err := h.service.GetWithReceptions(r.Context(), startDate, endDate, page, limit)
	if err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, "ошибка при получении списка ПВЗ")
		return
	}

	httpresponse.JSON(w, http.StatusOK, pvzs)
}

// CreatePVZ создает новый ПВЗ
func (h *PVZHandler) CreatePVZ(w http.ResponseWriter, r *http.Request) {
	var req struct {
		City string `json:"city"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Получаем ID модератора из контекста
	moderatorID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	moderatorUUID, err := uuid.Parse(moderatorID)
	if err != nil {
		http.Error(w, "invalid moderator ID", http.StatusBadRequest)
		return
	}

	// Создаем ПВЗ
	newPVZ, err := h.service.Create(r.Context(), req.City, moderatorUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Преобразуем в DTO
	response := struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		City      string `json:"city"`
	}{
		ID:        newPVZ.ID.String(),
		CreatedAt: newPVZ.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		City:      newPVZ.City,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetPVZ получает ПВЗ по ID
func (h *PVZHandler) GetPVZ(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	pvzID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid PVZ ID", http.StatusBadRequest)
		return
	}

	pvz, err := h.service.GetByID(r.Context(), pvzID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		City      string `json:"city"`
	}{
		ID:        pvz.ID.String(),
		CreatedAt: pvz.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		City:      pvz.City,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdatePVZ обновляет данные ПВЗ
func (h *PVZHandler) UpdatePVZ(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "invalid PVZ ID", http.StatusBadRequest)
		return
	}

	pvzID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid PVZ ID", http.StatusBadRequest)
		return
	}

	// Получаем ID модератора из контекста
	moderatorID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	moderatorUUID, err := uuid.Parse(moderatorID)
	if err != nil {
		http.Error(w, "invalid moderator ID", http.StatusBadRequest)
		return
	}

	var req struct {
		City string `json:"city"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Обновляем ПВЗ
	pvz := &domainPVZ.PVZ{
		ID:   pvzID,
		City: req.City,
	}

	if err := h.service.Update(r.Context(), pvz, moderatorUUID); err != nil {
		switch err {
		case servicePVZ.ErrPVZNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case servicePVZ.ErrAccessDenied:
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeletePVZ удаляет ПВЗ
func (h *PVZHandler) DeletePVZ(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "invalid PVZ ID", http.StatusBadRequest)
		return
	}

	pvzID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid PVZ ID", http.StatusBadRequest)
		return
	}

	// Получаем ID модератора из контекста
	moderatorID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	moderatorUUID, err := uuid.Parse(moderatorID)
	if err != nil {
		http.Error(w, "invalid moderator ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), pvzID, moderatorUUID); err != nil {
		switch err {
		case servicePVZ.ErrPVZNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case servicePVZ.ErrAccessDenied:
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListPVZ возвращает список ПВЗ
func (h *PVZHandler) ListPVZ(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if limit <= 0 {
		limit = 10
	}

	pvzs, err := h.service.List(r.Context(), offset, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		City      string `json:"city"`
	}, len(pvzs))

	for i, p := range pvzs {
		response[i] = struct {
			ID        string `json:"id"`
			CreatedAt string `json:"created_at"`
			City      string `json:"city"`
		}{
			ID:        p.ID.String(),
			CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			City:      p.City,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
