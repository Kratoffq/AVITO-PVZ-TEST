package pvz

import (
	"time"

	"github.com/google/uuid"
)

// PVZ представляет собой пункт выдачи заказов
type PVZ struct {
	ID        uuid.UUID `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	City      string    `db:"city"`
}

// New создает новый экземпляр PVZ
func New(city string) *PVZ {
	return &PVZ{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		City:      city,
	}
}
