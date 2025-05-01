package auth

import (
	"net/http"
	"strings"

	"github.com/avito/pvz/pkg/httpresponse"
)

// Middleware проверяет JWT токен в заголовке Authorization
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из заголовка
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httpresponse.Error(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		// Проверяем формат заголовка
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			httpresponse.Error(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		// Проверяем токен
		claims, err := ValidateToken(parts[1])
		if err != nil {
			httpresponse.Error(w, http.StatusUnauthorized, "invalid token")
			return
		}

		// Добавляем данные пользователя в контекст
		ctx := r.Context()
		ctx = WithUserID(ctx, claims.UserID)
		ctx = WithUserRole(ctx, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
