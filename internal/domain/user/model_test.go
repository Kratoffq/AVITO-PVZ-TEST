package user

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		role     Role
		want     *User
	}{
		{
			name:     "создание администратора",
			email:    "admin@example.com",
			password: "admin123",
			role:     RoleAdmin,
			want: &User{
				Email:    "admin@example.com",
				Password: "admin123",
				Role:     RoleAdmin,
			},
		},
		{
			name:     "создание сотрудника",
			email:    "employee@example.com",
			password: "emp123",
			role:     RoleEmployee,
			want: &User{
				Email:    "employee@example.com",
				Password: "emp123",
				Role:     RoleEmployee,
			},
		},
		{
			name:     "создание обычного пользователя",
			email:    "user@example.com",
			password: "user123",
			role:     RoleUser,
			want: &User{
				Email:    "user@example.com",
				Password: "user123",
				Role:     RoleUser,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.email, tt.password, tt.role)

			// Проверяем, что ID сгенерирован
			assert.NotEmpty(t, got.ID)

			// Проверяем, что время создания установлено корректно
			assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)

			// Проверяем остальные поля
			assert.Equal(t, tt.want.Email, got.Email)
			assert.Equal(t, tt.want.Password, got.Password)
			assert.Equal(t, tt.want.Role, got.Role)
		})
	}
}

func TestUser_CanCreatePVZ(t *testing.T) {
	tests := []struct {
		name string
		user *User
		want bool
	}{
		{
			name: "администратор может создавать ПВЗ",
			user: &User{Role: RoleAdmin},
			want: true,
		},
		{
			name: "сотрудник не может создавать ПВЗ",
			user: &User{Role: RoleEmployee},
			want: false,
		},
		{
			name: "обычный пользователь не может создавать ПВЗ",
			user: &User{Role: RoleUser},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.user.CanCreatePVZ()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRole_Constants(t *testing.T) {
	// Проверяем константы ролей
	assert.Equal(t, Role("admin"), RoleAdmin)
	assert.Equal(t, Role("employee"), RoleEmployee)
	assert.Equal(t, Role("user"), RoleUser)
}
