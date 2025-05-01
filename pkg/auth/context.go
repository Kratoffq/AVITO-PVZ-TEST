package auth

import (
	"context"

	"github.com/avito/pvz/internal/domain/user"
	"github.com/google/uuid"
)

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	userRoleKey contextKey = "user_role"
)

// WithUserID добавляет ID пользователя в контекст
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID получает ID пользователя из контекста
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}

// WithUserRole добавляет роль пользователя в контекст
func WithUserRole(ctx context.Context, role user.Role) context.Context {
	return context.WithValue(ctx, userRoleKey, role)
}

// GetUserRole получает роль пользователя из контекста
func GetUserRole(ctx context.Context) (user.Role, bool) {
	role, ok := ctx.Value(userRoleKey).(user.Role)
	return role, ok
}
