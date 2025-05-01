package pvz

import (
	"time"

	"github.com/avito/pvz/internal/domain/pvz"
	"github.com/google/uuid"
)

// CreatePVZRequest представляет запрос на создание ПВЗ
type CreatePVZRequest struct {
	City string
}

// CreatePVZResponse представляет ответ на создание ПВЗ
type CreatePVZResponse struct {
	ID        uuid.UUID `json:"id"`
	City      string    `json:"city"`
	CreatedAt time.Time `json:"created_at"`
}

// GetPVZWithReceptionsRequest представляет запрос на получение ПВЗ с приемками
type GetPVZWithReceptionsRequest struct {
	StartDate time.Time
	EndDate   time.Time
	Page      int
	Limit     int
}

// GetPVZWithReceptionsResponse представляет ответ с ПВЗ и приемками
type GetPVZWithReceptionsResponse struct {
	Items []*PVZWithReceptionsDTO `json:"items"`
}

// PVZWithReceptionsDTO представляет ПВЗ с приемками для ответа API
type PVZWithReceptionsDTO struct {
	ID         uuid.UUID       `json:"id"`
	City       string          `json:"city"`
	CreatedAt  time.Time       `json:"created_at"`
	Receptions []*ReceptionDTO `json:"receptions"`
}

// ReceptionDTO представляет приемку для ответа API
type ReceptionDTO struct {
	ID       uuid.UUID     `json:"id"`
	DateTime time.Time     `json:"date_time"`
	Status   string        `json:"status"`
	Products []*ProductDTO `json:"products"`
}

// ProductDTO представляет товар для ответа API
type ProductDTO struct {
	ID       uuid.UUID `json:"id"`
	DateTime time.Time `json:"date_time"`
	Type     string    `json:"type"`
}

// ToPVZResponse преобразует доменную модель в DTO
func ToPVZResponse(p *pvz.PVZ) *CreatePVZResponse {
	if p == nil {
		return nil
	}
	return &CreatePVZResponse{
		ID:        p.ID,
		City:      p.City,
		CreatedAt: p.CreatedAt,
	}
}

// ToPVZWithReceptionsResponse преобразует доменные модели в DTO
func ToPVZWithReceptionsResponse(items []*pvz.PVZWithReceptions) *GetPVZWithReceptionsResponse {
	if items == nil {
		return &GetPVZWithReceptionsResponse{
			Items: make([]*PVZWithReceptionsDTO, 0),
		}
	}

	result := &GetPVZWithReceptionsResponse{
		Items: make([]*PVZWithReceptionsDTO, len(items)),
	}

	for i, item := range items {
		if item == nil || item.PVZ == nil {
			continue
		}

		result.Items[i] = &PVZWithReceptionsDTO{
			ID:         item.PVZ.ID,
			City:       item.PVZ.City,
			CreatedAt:  item.PVZ.CreatedAt,
			Receptions: make([]*ReceptionDTO, len(item.Receptions)),
		}

		for j, r := range item.Receptions {
			if r == nil || r.Reception == nil {
				continue
			}

			result.Items[i].Receptions[j] = &ReceptionDTO{
				ID:       r.Reception.ID,
				DateTime: r.Reception.DateTime,
				Status:   string(r.Reception.Status),
				Products: make([]*ProductDTO, len(r.Products)),
			}

			for k, p := range r.Products {
				if p == nil {
					continue
				}

				result.Items[i].Receptions[j].Products[k] = &ProductDTO{
					ID:       p.ID,
					DateTime: p.DateTime,
					Type:     string(p.Type),
				}
			}
		}
	}

	return result
}
