package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	user := User{
		ID:        id,
		Email:     "test@example.com",
		Password:  "password123",
		Role:      RoleEmployee,
		CreatedAt: now,
	}

	assert.Equal(t, id, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "password123", user.Password)
	assert.Equal(t, RoleEmployee, user.Role)
	assert.Equal(t, now, user.CreatedAt)
}

func TestPVZ(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	pvz := PVZ{
		ID:               id,
		RegistrationDate: now,
		City:             "Москва",
	}

	assert.Equal(t, id, pvz.ID)
	assert.Equal(t, now, pvz.RegistrationDate)
	assert.Equal(t, "Москва", pvz.City)
}

func TestReception(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	pvzID := uuid.New()
	reception := Reception{
		ID:       id,
		DateTime: now,
		PVZID:    pvzID,
		Status:   StatusInProgress,
	}

	assert.Equal(t, id, reception.ID)
	assert.Equal(t, now, reception.DateTime)
	assert.Equal(t, pvzID, reception.PVZID)
	assert.Equal(t, StatusInProgress, reception.Status)
}

func TestProduct(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	receptionID := uuid.New()
	product := Product{
		ID:          id,
		DateTime:    now,
		Type:        TypeElectronics,
		ReceptionID: receptionID,
	}

	assert.Equal(t, id, product.ID)
	assert.Equal(t, now, product.DateTime)
	assert.Equal(t, TypeElectronics, product.Type)
	assert.Equal(t, receptionID, product.ReceptionID)
}

func TestReceptionStatus(t *testing.T) {
	tests := []struct {
		name   string
		status ReceptionStatus
	}{
		{
			name:   "статус в процессе",
			status: StatusInProgress,
		},
		{
			name:   "статус закрыт",
			status: StatusClose,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.status)
			assert.IsType(t, ReceptionStatus(""), tt.status)
		})
	}
}

func TestProductType(t *testing.T) {
	tests := []struct {
		name string
		typ  ProductType
	}{
		{
			name: "тип электроника",
			typ:  TypeElectronics,
		},
		{
			name: "тип одежда",
			typ:  TypeClothing,
		},
		{
			name: "тип еда",
			typ:  TypeFood,
		},
		{
			name: "тип другое",
			typ:  TypeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.typ)
			assert.IsType(t, ProductType(""), tt.typ)
		})
	}
}

func TestUserRole(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
	}{
		{
			name: "роль сотрудник",
			role: RoleEmployee,
		},
		{
			name: "роль модератор",
			role: RoleModerator,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.role)
			assert.IsType(t, UserRole(""), tt.role)
		})
	}
}
