package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleEmployee  UserRole = "employee"
	RoleModerator UserRole = "moderator"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	Role      UserRole  `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type PVZ struct {
	ID               uuid.UUID `json:"id" db:"id"`
	RegistrationDate time.Time `json:"registration_date" db:"registration_date"`
	City             string    `json:"city" db:"city"`
}

type ReceptionStatus string

const (
	StatusInProgress ReceptionStatus = "in_progress"
	StatusClose      ReceptionStatus = "close"
)

type Reception struct {
	ID       uuid.UUID       `json:"id" db:"id"`
	DateTime time.Time       `json:"date_time" db:"date_time"`
	PVZID    uuid.UUID       `json:"pvz_id" db:"pvz_id"`
	Status   ReceptionStatus `json:"status" db:"status"`
}

type ProductType string

const (
	TypeElectronics ProductType = "электроника"
	TypeClothing    ProductType = "одежда"
	TypeShoes       ProductType = "обувь"
)

type Product struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	DateTime    time.Time   `json:"date_time" db:"date_time"`
	Type        ProductType `json:"type" db:"type"`
	ReceptionID uuid.UUID   `json:"reception_id" db:"reception_id"`
}
