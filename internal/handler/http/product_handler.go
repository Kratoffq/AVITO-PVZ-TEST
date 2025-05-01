package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/avito/pvz/internal/domain/product"
	domainUser "github.com/avito/pvz/internal/domain/user"
	"github.com/avito/pvz/internal/handler/http/middleware"
	productService "github.com/avito/pvz/internal/service/product"
	"github.com/avito/pvz/pkg/httpresponse"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ProductHandler обрабатывает HTTP-запросы для продуктов
type ProductHandler struct {
	service *productService.Service
}

// NewProductHandler создает новый экземпляр ProductHandler
func NewProductHandler(service *productService.Service) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

// RegisterRoutes регистрирует маршруты для продуктов
func (h *ProductHandler) RegisterRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Use(middleware.RequireRole(domainUser.RoleEmployee))

		r.Post("/product", h.Create)
		r.Post("/product/batch", h.CreateBatch)
		r.Delete("/product/last/{reception_id}", h.DeleteLast)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Get("/product/{id}", h.GetByID)
		r.Get("/product/reception/{reception_id}", h.GetByReceptionID)
		r.Get("/product", h.List)
	})
}

// Create обрабатывает создание продукта
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ReceptionID string       `json:"reception_id"`
		Type        product.Type `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	receptionID, err := uuid.Parse(req.ReceptionID)
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат ID приемки")
		return
	}

	product, err := h.service.Create(r.Context(), receptionID, req.Type)
	if err != nil {
		switch err {
		case productService.ErrReceptionNotFound:
			httpresponse.Error(w, http.StatusNotFound, "приемка не найдена")
		case productService.ErrReceptionAlreadyClose:
			httpresponse.Error(w, http.StatusBadRequest, "приемка уже закрыта")
		case productService.ErrInvalidProductType:
			httpresponse.Error(w, http.StatusBadRequest, "неверный тип товара")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "ошибка при добавлении товара")
		}
		return
	}

	httpresponse.JSON(w, http.StatusCreated, product)
}

// CreateBatch обрабатывает создание нескольких продуктов
func (h *ProductHandler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ReceptionID string         `json:"reception_id"`
		Types       []product.Type `json:"types"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	receptionID, err := uuid.Parse(req.ReceptionID)
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат ID приемки")
		return
	}

	if err := h.service.CreateBatch(r.Context(), receptionID, req.Types); err != nil {
		switch err {
		case productService.ErrReceptionNotFound:
			httpresponse.Error(w, http.StatusNotFound, "приемка не найдена")
		case productService.ErrReceptionAlreadyClose:
			httpresponse.Error(w, http.StatusBadRequest, "приемка уже закрыта")
		case productService.ErrInvalidProductType:
			httpresponse.Error(w, http.StatusBadRequest, "неверный тип товара")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "ошибка при добавлении товаров")
		}
		return
	}

	httpresponse.JSON(w, http.StatusCreated, nil)
}

// DeleteLast обрабатывает удаление последнего продукта
func (h *ProductHandler) DeleteLast(w http.ResponseWriter, r *http.Request) {
	receptionID, err := uuid.Parse(chi.URLParam(r, "reception_id"))
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат ID приемки")
		return
	}

	if err := h.service.DeleteLast(r.Context(), receptionID); err != nil {
		switch err {
		case productService.ErrReceptionNotFound:
			httpresponse.Error(w, http.StatusNotFound, "приемка не найдена")
		case productService.ErrReceptionAlreadyClose:
			httpresponse.Error(w, http.StatusBadRequest, "приемка уже закрыта")
		case productService.ErrProductNotFound:
			httpresponse.Error(w, http.StatusNotFound, "товар не найден")
		default:
			httpresponse.Error(w, http.StatusInternalServerError, "ошибка при удалении товара")
		}
		return
	}

	httpresponse.JSON(w, http.StatusOK, nil)
}

// GetByID получает товар по ID
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	productID, err := uuid.Parse(id)
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат ID товара")
		return
	}

	p, err := h.service.GetByID(r.Context(), productID)
	if err != nil {
		httpresponse.Error(w, http.StatusNotFound, "товар не найден")
		return
	}

	httpresponse.JSON(w, http.StatusOK, p)
}

// GetByReceptionID получает все товары приемки
func (h *ProductHandler) GetByReceptionID(w http.ResponseWriter, r *http.Request) {
	receptionID := chi.URLParam(r, "reception_id")
	receptionUUID, err := uuid.Parse(receptionID)
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "неверный формат ID приемки")
		return
	}

	products, err := h.service.GetByReceptionID(r.Context(), receptionUUID)
	if err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, "ошибка при получении товаров")
		return
	}

	httpresponse.JSON(w, http.StatusOK, products)
}

// List возвращает список товаров
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if limit <= 0 {
		limit = 10
	}

	products, err := h.service.List(r.Context(), offset, limit)
	if err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, "ошибка при получении списка товаров")
		return
	}

	httpresponse.JSON(w, http.StatusOK, products)
}
