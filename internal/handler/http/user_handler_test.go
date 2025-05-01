package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainUser "github.com/avito/pvz/internal/domain/user"
	userService "github.com/avito/pvz/internal/service/user"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserService реализует мок для user.ServiceInterface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, email, password string, role domainUser.Role) (*domainUser.User, error) {
	args := m.Called(ctx, email, password, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, email, password string) (*domainUser.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserService) GetByID(ctx context.Context, id uuid.UUID) (*domainUser.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserService) Update(ctx context.Context, user *domainUser.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) List(ctx context.Context, offset, limit int) ([]*domainUser.User, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*domainUser.User), args.Error(1)
}

func (m *MockUserService) LoginUser(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func TestUserHandler_RegisterUser(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "успешная регистрация",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "testpass",
				"role":     "moderator",
			},
			mockSetup: func(m *MockUserService) {
				m.On("Register", mock.Anything, "test@example.com", "testpass", domainUser.Role("moderator")).Return(&domainUser.User{
					ID:    uuid.New(),
					Email: "test@example.com",
					Role:  domainUser.Role("moderator"),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":    mock.Anything,
				"email": "test@example.com",
				"role":  "moderator",
			},
		},
		{
			name:        "неверный формат тела запроса",
			requestBody: "invalid json",
			mockSetup: func(m *MockUserService) {
				// Для неверного формата тела запроса мок не нужен, так как до вызова сервиса не дойдет
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name: "пустой email",
			requestBody: map[string]interface{}{
				"email":    "",
				"password": "testpass",
				"role":     "moderator",
			},
			mockSetup: func(m *MockUserService) {
				m.On("Register", mock.Anything, "", "testpass", domainUser.Role("moderator")).Return(nil, userService.ErrInvalidEmail)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "email cannot be empty",
		},
		{
			name: "пустой пароль",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "",
				"role":     "moderator",
			},
			mockSetup: func(m *MockUserService) {
				m.On("Register", mock.Anything, "test@example.com", "", domainUser.Role("moderator")).Return(nil, userService.ErrInvalidPassword)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "password cannot be empty",
		},
		{
			name: "неверная роль",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "testpass",
				"role":     "invalid",
			},
			mockSetup: func(m *MockUserService) {
				m.On("Register", mock.Anything, "test@example.com", "testpass", domainUser.Role("invalid")).Return(nil, userService.ErrInvalidRole)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid role",
		},
		{
			name: "ошибка сервиса",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "testpass",
				"role":     "moderator",
			},
			mockSetup: func(m *MockUserService) {
				m.On("Register", mock.Anything, "test@example.com", "testpass", domainUser.Role("moderator")).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.RegisterUser(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err)
				for key, value := range tt.expectedBody.(map[string]interface{}) {
					if value == mock.Anything {
						assert.NotEmpty(t, response[key])
					} else {
						assert.Equal(t, value, response[key])
					}
				}
			} else {
				assert.Contains(t, rr.Body.String(), tt.expectedBody.(string))
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_LoginUser(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "успешный вход",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockSetup: func(m *MockUserService) {
				userID := uuid.New()
				m.On("Login", mock.Anything, "test@example.com", "password123").
					Return(&domainUser.User{
						ID:        userID,
						Email:     "test@example.com",
						Role:      domainUser.RoleUser,
						CreatedAt: time.Now(),
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"email": "test@example.com",
				"role":  "user",
			},
		},
		{
			name: "неверные учетные данные",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			mockSetup: func(m *MockUserService) {
				m.On("Login", mock.Anything, "test@example.com", "wrongpassword").
					Return(nil, userService.ErrInvalidPassword)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "неверные учетные данные",
			},
		},
		{
			name:           "неверный формат тела запроса",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "неверный формат запроса",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.LoginUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "успешное получение пользователя",
			userID: userID.String(),
			mockSetup: func(m *MockUserService) {
				m.On("GetByID", mock.Anything, userID).
					Return(&domainUser.User{
						ID:        userID,
						Email:     "test@example.com",
						Role:      domainUser.RoleUser,
						CreatedAt: time.Now(),
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"email": "test@example.com",
				"role":  "user",
			},
		},
		{
			name:   "неверный формат ID",
			userID: "invalid-uuid",
			mockSetup: func(m *MockUserService) {
				// Не ожидаем вызовов сервиса при неверном формате ID
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
		{
			name:   "пользователь не найден",
			userID: userID.String(),
			mockSetup: func(m *MockUserService) {
				m.On("GetByID", mock.Anything, userID).
					Return(nil, userService.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
			w := httptest.NewRecorder()

			handler.GetUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody["email"], response["email"])
				assert.Equal(t, tt.expectedBody["role"], response["role"])
			}
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name           string
		userID         string
		requestBody    map[string]interface{}
		mockSetup      func(*MockUserService)
		expectedStatus int
	}{
		{
			name:   "успешное обновление пользователя",
			userID: userID.String(),
			requestBody: map[string]interface{}{
				"email": "updated@example.com",
				"role":  "admin",
			},
			mockSetup: func(m *MockUserService) {
				m.On("Update", mock.Anything, &domainUser.User{
					ID:    userID,
					Email: "updated@example.com",
					Role:  domainUser.RoleAdmin,
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "неверный формат ID",
			userID:         "invalid-uuid",
			requestBody:    map[string]interface{}{},
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "неверный формат тела запроса",
			userID: userID.String(),
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "пользователь не найден",
			userID: userID.String(),
			requestBody: map[string]interface{}{
				"email": "updated@example.com",
				"role":  "admin",
			},
			mockSetup: func(m *MockUserService) {
				m.On("Update", mock.Anything, &domainUser.User{
					ID:    userID,
					Email: "updated@example.com",
					Role:  domainUser.RoleAdmin,
				}).Return(userService.ErrUserNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.userID, bytes.NewBuffer(body))
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
			w := httptest.NewRecorder()

			handler.UpdateUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockUserService)
		expectedStatus int
	}{
		{
			name:   "успешное удаление пользователя",
			userID: userID.String(),
			mockSetup: func(m *MockUserService) {
				m.On("Delete", mock.Anything, userID).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "неверный формат ID",
			userID:         "invalid-uuid",
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "пользователь не найден",
			userID: userID.String(),
			mockSetup: func(m *MockUserService) {
				m.On("Delete", mock.Anything, userID).Return(userService.ErrUserNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.userID, nil)
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
			w := httptest.NewRecorder()

			handler.DeleteUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:        "успешное получение списка пользователей",
			queryParams: "?offset=0&limit=10",
			mockSetup: func(m *MockUserService) {
				users := []*domainUser.User{
					{
						ID:        uuid.New(),
						Email:     "user1@example.com",
						Role:      domainUser.RoleUser,
						CreatedAt: time.Now(),
					},
					{
						ID:        uuid.New(),
						Email:     "user2@example.com",
						Role:      domainUser.RoleAdmin,
						CreatedAt: time.Now(),
					},
				}
				m.On("List", mock.Anything, 0, 10).Return(users, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:        "получение списка с неверными параметрами",
			queryParams: "?offset=invalid&limit=invalid",
			mockSetup: func(m *MockUserService) {
				users := []*domainUser.User{}
				m.On("List", mock.Anything, 0, 10).Return(users, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:        "ошибка сервиса",
			queryParams: "?offset=0&limit=10",
			mockSetup: func(m *MockUserService) {
				m.On("List", mock.Anything, 0, 10).Return([]*domainUser.User(nil), errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/users"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ListUsers(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCount > 0 {
				var response []map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(response))
			}

			mockService.AssertExpectations(t)
		})
	}
}
