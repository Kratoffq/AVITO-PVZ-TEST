package models

import (
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя системы
type User struct {
	ID        uuid.UUID
	Email     string
	Password  string
	Role      UserRole
	CreatedAt time.Time
}

// PVZ представляет пункт выдачи заказов
type PVZ struct {
	ID               uuid.UUID
	RegistrationDate time.Time
	City             string
}

// Reception представляет приемку товаров
type Reception struct {
	ID       uuid.UUID
	DateTime time.Time
	PVZID    uuid.UUID
	Status   ReceptionStatus
}

// ReceptionStatus представляет статус приемки
type ReceptionStatus string

const (
	StatusInProgress ReceptionStatus = "in_progress"
	StatusClose      ReceptionStatus = "close"
)

// Product представляет товар
type Product struct {
	ID          uuid.UUID
	DateTime    time.Time
	Type        ProductType
	ReceptionID uuid.UUID
}

// ProductType представляет тип товара
type ProductType string

const (
	TypeElectronics ProductType = "electronics"
	TypeClothing    ProductType = "clothing"
	TypeFood        ProductType = "food"
	TypeOther       ProductType = "other"
)

// UserRole определяет роль пользователя
type UserRole string

const (
	RoleEmployee  UserRole = "employee"
	RoleModerator UserRole = "moderator"
)
