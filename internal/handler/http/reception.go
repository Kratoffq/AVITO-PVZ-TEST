package http

import (
	"encoding/json"
	"net/http"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/avito/pvz/internal/service/reception"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// RegisterReceptionHandlers регистрирует HTTP обработчики для приёмок
func RegisterReceptionHandlers(router *mux.Router, service *reception.Service) {
	router.HandleFunc("/api/v1/reception", createReception(service)).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/reception/{id}", getReception(service)).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/reception/{id}/product", addProduct(service)).Methods(http.MethodPost)
}

func createReception(service *reception.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			PVZID string `json:"pvz_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		pvzID, err := uuid.Parse(req.PVZID)
		if err != nil {
			http.Error(w, "Invalid PVZ ID", http.StatusBadRequest)
			return
		}

		rec, err := service.Create(r.Context(), pvzID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rec)
	}
}

func getReception(service *reception.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "Invalid reception ID", http.StatusBadRequest)
			return
		}

		rec, err := service.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rec)
	}
}

func addProduct(service *reception.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		receptionID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "Invalid reception ID", http.StatusBadRequest)
			return
		}

		var req struct {
			Type string `json:"type"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Проверяем тип товара
		switch req.Type {
		case string(product.TypeElectronics),
			string(product.TypeClothing),
			string(product.TypeFood),
			string(product.TypeOther):
		default:
			http.Error(w, "Invalid product type", http.StatusBadRequest)
			return
		}

		err = service.CreateProduct(r.Context(), receptionID, req.Type)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
