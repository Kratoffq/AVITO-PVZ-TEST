package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/avito/pvz/internal/domain/user"
	"github.com/avito/pvz/pkg/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "отсутствует заголовок авторизации",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   ErrNoAuthHeader.Error(),
		},
		{
			name:           "неверный формат заголовка",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   ErrInvalidAuthHeader.Error(),
		},
		{
			name:           "неверный тип токена",
			authHeader:     "Basic token123",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   ErrInvalidAuthHeader.Error(),
		},
		{
			name:           "неверный токен",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   ErrInvalidToken.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}

	t.Run("успешная авторизация", func(t *testing.T) {
		// Создаем валидный токен
		userID := uuid.MustParse("c54e392f-75b1-4e33-9858-e1810bd9549f")
		token, err := auth.GenerateToken(userID, user.RoleAdmin)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		rr := httptest.NewRecorder()
		handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := GetUserID(r.Context())
			require.NoError(t, err)
			assert.Equal(t, userID.String(), id)

			role, err := GetUserRole(r.Context())
			require.NoError(t, err)
			assert.Equal(t, user.RoleAdmin, role)

			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		userRole       user.Role
		requiredRole   user.Role
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "доступ разрешен",
			userRole:       user.RoleAdmin,
			requiredRole:   user.RoleAdmin,
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "доступ запрещен",
			userRole:       user.RoleUser,
			requiredRole:   user.RoleAdmin,
			expectedStatus: http.StatusForbidden,
			expectedBody:   ErrAccessDenied.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := context.WithValue(req.Context(), UserRoleKey, tt.userRole)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := RequireRole(tt.requiredRole)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		expectedID  string
		expectedErr error
	}{
		{
			name: "успешное получение ID",
			ctx: func() context.Context {
				id := "c54e392f-75b1-4e33-9858-e1810bd9549f"
				return context.WithValue(context.Background(), UserIDKey, id)
			}(),
			expectedID:  "c54e392f-75b1-4e33-9858-e1810bd9549f",
			expectedErr: nil,
		},
		{
			name:        "ID отсутствует в контексте",
			ctx:         context.Background(),
			expectedID:  "",
			expectedErr: ErrInvalidToken,
		},
		{
			name:        "неверный тип ID",
			ctx:         context.WithValue(context.Background(), UserIDKey, 123),
			expectedID:  "",
			expectedErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := GetUserID(tt.ctx)
			assert.Equal(t, tt.expectedID, id)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestGetUserRole(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		expectedRole user.Role
		expectedErr  error
	}{
		{
			name:         "успешное получение роли",
			ctx:          context.WithValue(context.Background(), UserRoleKey, user.RoleAdmin),
			expectedRole: user.RoleAdmin,
			expectedErr:  nil,
		},
		{
			name:         "роль отсутствует в контексте",
			ctx:          context.Background(),
			expectedRole: "",
			expectedErr:  ErrInvalidToken,
		},
		{
			name:         "неверный тип роли",
			ctx:          context.WithValue(context.Background(), UserRoleKey, 123),
			expectedRole: "",
			expectedErr:  ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := GetUserRole(tt.ctx)
			assert.Equal(t, tt.expectedRole, role)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
