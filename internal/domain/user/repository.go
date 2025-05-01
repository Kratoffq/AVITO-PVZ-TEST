package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Repository определяет методы для работы с пользователями в хранилище
type Repository interface {
	// Create создает нового пользователя
	Create(ctx context.Context, user *User) error

	// GetByID получает пользователя по ID
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetByEmail получает пользователя по email
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Update обновляет данные пользователя
	Update(ctx context.Context, user *User) error

	// Delete удаляет пользователя по ID
	Delete(ctx context.Context, id uuid.UUID) error

	// List возвращает список пользователей с пагинацией
	List(ctx context.Context, offset, limit int) ([]*User, error)
}

// ErrUserNotFound возвращается, когда пользователь не найден
var ErrUserNotFound = errors.New("user not found")

// ErrUserAlreadyExists возвращается при попытке создать пользователя с существующим email
var ErrUserAlreadyExists = errors.New("user already exists")
