package pvz

import (
	"time"
)

// PVZ представляет собой модель пункта выдачи заказов
type PVZ struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Schedule  Schedule  `json:"schedule"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Schedule представляет график работы PVZ
type Schedule struct {
	OpenTime  string `json:"open_time"`
	CloseTime string `json:"close_time"`
	Weekend   bool   `json:"weekend"`
}

// Status представляет статус PVZ
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusDeleted  Status = "deleted"
)

// Reception представляет информацию о приеме товара в PVZ
type Reception struct {
	ID        int64     `json:"id"`
	PVZID     int64     `json:"pvz_id"`
	ProductID int64     `json:"product_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
