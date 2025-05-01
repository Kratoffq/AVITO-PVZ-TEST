package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/avito/pvz/internal/domain/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	userID := uuid.New()
	role := user.Role("admin")

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
			expectedBody:   `{"error":"missing authorization header"}`,
		},
		{
			name:           "неверный формат заголовка",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid authorization header format"}`,
		},
		{
			name:           "неверный формат токена",
			authHeader:     "Bearer invalid.token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый обработчик
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Проверяем, что данные пользователя добавлены в контекст
				ctxUserID, ok := GetUserID(r.Context())
				assert.True(t, ok)
				assert.Equal(t, userID, ctxUserID)

				ctxRole, ok := GetUserRole(r.Context())
				assert.True(t, ok)
				assert.Equal(t, role, ctxRole)

				w.WriteHeader(http.StatusOK)
			})

			// Создаем тестовый запрос
			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Создаем ResponseRecorder для записи ответа
			rr := httptest.NewRecorder()

			// Создаем middleware с тестовым обработчиком
			middleware := Middleware(handler)

			// Выполняем запрос
			middleware.ServeHTTP(rr, req)

			// Проверяем статус и тело ответа
			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			}
		})
	}

	// Тест с валидным токеном
	t.Run("валидный токен", func(t *testing.T) {
		// Генерируем валидный токен
		token, err := GenerateToken(userID, role)
		assert.NoError(t, err)

		// Создаем тестовый обработчик
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, что данные пользователя добавлены в контекст
			ctxUserID, ok := GetUserID(r.Context())
			assert.True(t, ok)
			assert.Equal(t, userID, ctxUserID)

			ctxRole, ok := GetUserRole(r.Context())
			assert.True(t, ok)
			assert.Equal(t, role, ctxRole)

			w.WriteHeader(http.StatusOK)
		})

		// Создаем тестовый запрос с валидным токеном
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Создаем middleware с тестовым обработчиком
		middleware := Middleware(handler)

		// Выполняем запрос
		middleware.ServeHTTP(rr, req)

		// Проверяем статус
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
