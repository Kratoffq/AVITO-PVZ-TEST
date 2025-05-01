package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	receptionService "github.com/avito/pvz/internal/service/reception"
	"github.com/avito/pvz/pkg/httpresponse"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ReceptionHandler обрабатывает HTTP-запросы для приемок
type ReceptionHandler struct {
	service ReceptionServiceInterface
}

// NewReceptionHandler создает новый экземпляр ReceptionHandler
func NewReceptionHandler(service ReceptionServiceInterface) *ReceptionHandler {
	return &ReceptionHandler{
		service: service,
	}
}

// RegisterRoutes регистрирует маршруты для приемок
func (h *ReceptionHandler) RegisterRoutes(r chi.Router) {
	r.Post("/reception", h.Create)
	r.Get("/reception/{id}", h.GetByID)
	r.Post("/reception/close", h.Close)
}

// Create обрабатывает создание приемки
func (h *ReceptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PVZID uuid.UUID `json:"pvz_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	reception, err := h.service.Create(r.Context(), req.PVZID)
	if err != nil {
		switch err {
		case receptionService.ErrPVZNotFound:
			httpresponse.Error(w, http.StatusNotFound, "ПВЗ не найден")
		case receptionService.ErrReceptionAlreadyOpen:
			httpresponse.Error(w, http.StatusBadRequest, "у ПВЗ уже есть открытая приемка")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "ошибка при создании приемки")
		}
		return
	}

	httpresponse.JSON(w, http.StatusCreated, reception)
}

// GetByID обрабатывает получение приемки по ID
func (h *ReceptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат ID приемки")
		return
	}

	reception, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		httpresponse.Error(w, http.StatusNotFound, "приемка не найдена")
		return
	}

	httpresponse.JSON(w, http.StatusOK, reception)
}

// Close обрабатывает закрытие приемки
func (h *ReceptionHandler) Close(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PVZID string `json:"pvz_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	pvzID, err := uuid.Parse(req.PVZID)
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат ID ПВЗ")
		return
	}

	if err := h.service.Close(r.Context(), pvzID); err != nil {
		switch err {
		case receptionService.ErrPVZNotFound:
			httpresponse.Error(w, http.StatusNotFound, "ПВЗ не найден")
		case receptionService.ErrReceptionNotFound:
			httpresponse.Error(w, http.StatusNotFound, "приемка не найдена")
		case receptionService.ErrReceptionAlreadyClose:
			httpresponse.Error(w, http.StatusBadRequest, "приемка уже закрыта")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "ошибка при закрытии приемки")
		}
		return
	}

	httpresponse.JSON(w, http.StatusOK, nil)
}

// GetReception получает приемку по ID
func (h *ReceptionHandler) GetReception(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	receptionID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid reception ID", http.StatusBadRequest)
		return
	}

	reception, err := h.service.GetByID(r.Context(), receptionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := struct {
		ID       string `json:"id"`
		DateTime string `json:"date_time"`
		PVZID    string `json:"pvz_id"`
		Status   string `json:"status"`
	}{
		ID:       reception.ID.String(),
		DateTime: reception.DateTime.Format("2006-01-02T15:04:05Z07:00"),
		PVZID:    reception.PVZID.String(),
		Status:   string(reception.Status),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOpenReception получает открытую приемку для ПВЗ
func (h *ReceptionHandler) GetOpenReception(w http.ResponseWriter, r *http.Request) {
	pvzID := chi.URLParam(r, "pvz_id")
	pvzUUID, err := uuid.Parse(pvzID)
	if err != nil {
		http.Error(w, "invalid PVZ ID", http.StatusBadRequest)
		return
	}

	reception, err := h.service.GetOpenByPVZID(r.Context(), pvzUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := struct {
		ID       string `json:"id"`
		DateTime string `json:"date_time"`
		PVZID    string `json:"pvz_id"`
		Status   string `json:"status"`
	}{
		ID:       reception.ID.String(),
		DateTime: reception.DateTime.Format("2006-01-02T15:04:05Z07:00"),
		PVZID:    reception.PVZID.String(),
		Status:   string(reception.Status),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListReceptions возвращает список приемок
func (h *ReceptionHandler) ListReceptions(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if limit <= 0 {
		limit = 10
	}

	receptions, err := h.service.List(r.Context(), offset, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]struct {
		ID       string `json:"id"`
		DateTime string `json:"date_time"`
		PVZID    string `json:"pvz_id"`
		Status   string `json:"status"`
	}, len(receptions))

	for i, r := range receptions {
		response[i] = struct {
			ID       string `json:"id"`
			DateTime string `json:"date_time"`
			PVZID    string `json:"pvz_id"`
			Status   string `json:"status"`
		}{
			ID:       r.ID.String(),
			DateTime: r.DateTime.Format("2006-01-02T15:04:05Z07:00"),
			PVZID:    r.PVZID.String(),
			Status:   string(r.Status),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
