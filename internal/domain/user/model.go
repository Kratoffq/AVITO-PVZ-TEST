package user

import (
	"time"

	"github.com/google/uuid"
)

// Role представляет роль пользователя
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleEmployee Role = "employee"
	RoleUser     Role = "user"
)

// User представляет собой пользователя системы
type User struct {
	ID        uuid.UUID `db:"id"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	Role      Role      `db:"role"`
	CreatedAt time.Time `db:"created_at"`
}

// New создает нового пользователя
func New(email, password string, role Role) *User {
	return &User{
		ID:        uuid.New(),
		Email:     email,
		Password:  password,
		Role:      role,
		CreatedAt: time.Now(),
	}
}

// CanCreatePVZ проверяет, может ли пользователь создавать ПВЗ
func (u *User) CanCreatePVZ() bool {
	return u.Role == RoleAdmin
}
