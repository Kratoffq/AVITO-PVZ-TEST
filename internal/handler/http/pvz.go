package http

import (
	"encoding/json"
	"errors"
	"net/http"

	domainpvz "github.com/avito/pvz/internal/domain/pvz"
	pvzservice "github.com/avito/pvz/internal/service/pvz"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// RegisterPVZHandlers регистрирует HTTP обработчики для ПВЗ
func RegisterPVZHandlers(router *mux.Router, service *pvzservice.Service) {
	router.HandleFunc("/api/v1/pvz", createPVZ(service)).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/pvz/{id}", getPVZ(service)).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/pvz", getAllPVZ(service)).Methods(http.MethodGet)
}

func createPVZ(service *pvzservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			City string `json:"city"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.City == "" {
			http.Error(w, "City is required", http.StatusBadRequest)
			return
		}

		moderatorID, err := getModeratorID(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		pvz, err := service.Create(r.Context(), req.City, moderatorID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(pvz)
	}
}

func getPVZ(service *pvzservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "Invalid PVZ ID", http.StatusBadRequest)
			return
		}

		pvz, err := service.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, domainpvz.ErrNotFound) {
				http.Error(w, "PVZ not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pvz)
	}
}

func getAllPVZ(service *pvzservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pvzs, err := service.GetAll(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pvzs)
	}
}

func getModeratorID(r *http.Request) (uuid.UUID, error) {
	moderatorIDStr := r.Header.Get("X-Moderator-ID")
	if moderatorIDStr == "" {
		return uuid.Nil, http.ErrNoCookie
	}
	return uuid.Parse(moderatorIDStr)
}
