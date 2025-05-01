package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/avito/pvz/internal/domain/user"
	"github.com/avito/pvz/pkg/auth"
)

var (
	ErrNoAuthHeader      = errors.New("no authorization header")
	ErrInvalidAuthHeader = errors.New("invalid authorization header")
	ErrInvalidToken      = errors.New("invalid token")
	ErrAccessDenied      = errors.New("access denied")
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
)

// AuthMiddleware проверяет JWT токен и добавляет информацию о пользователе в контекст
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из заголовка
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, ErrNoAuthHeader.Error(), http.StatusUnauthorized)
			return
		}

		// Проверяем формат заголовка
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, ErrInvalidAuthHeader.Error(), http.StatusUnauthorized)
			return
		}

		// Проверяем токен
		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, ErrInvalidToken.Error(), http.StatusUnauthorized)
			return
		}

		// Добавляем информацию в контекст
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID.String())
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole проверяет, что у пользователя есть необходимая роль
func RequireRole(role user.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(UserRoleKey).(user.Role)
			if !ok || userRole != role {
				http.Error(w, ErrAccessDenied.Error(), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID получает ID пользователя из контекста
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return "", ErrInvalidToken
	}
	return userID, nil
}

// GetUserRole получает роль пользователя из контекста
func GetUserRole(ctx context.Context) (user.Role, error) {
	userRole, ok := ctx.Value(UserRoleKey).(user.Role)
	if !ok {
		return "", ErrInvalidToken
	}
	return userRole, nil
}
