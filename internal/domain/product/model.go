package product

import (
	"time"

	"github.com/google/uuid"
)

// Type представляет тип товара
type Type string

const (
	TypeElectronics Type = "electronics"
	TypeClothing    Type = "clothing"
	TypeFood        Type = "food"
	TypeOther       Type = "other"
)

// Product представляет собой товар
type Product struct {
	ID          uuid.UUID `db:"id"`
	DateTime    time.Time `db:"date_time"`
	Type        Type      `db:"type"`
	ReceptionID uuid.UUID `db:"reception_id"`
}

// New создает новый экземпляр Product
func New(receptionID uuid.UUID, productType Type) *Product {
	return &Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        productType,
		ReceptionID: receptionID,
	}
}
