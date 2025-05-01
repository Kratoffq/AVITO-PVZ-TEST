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

	domainPVZ "github.com/avito/pvz/internal/domain/pvz"
	"github.com/avito/pvz/internal/handler/http/middleware"
	servicePVZ "github.com/avito/pvz/internal/service/pvz"
	"github.com/avito/pvz/pkg/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockPVZService - мок для сервиса PVZ
type MockPVZService struct {
	mock.Mock
}

func (m *MockPVZService) Create(ctx context.Context, city string, userID uuid.UUID) (*domainPVZ.PVZ, error) {
	args := m.Called(ctx, city, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainPVZ.PVZ), args.Error(1)
}

func (m *MockPVZService) GetByID(ctx context.Context, id uuid.UUID) (*domainPVZ.PVZ, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainPVZ.PVZ), args.Error(1)
}

func (m *MockPVZService) GetWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*domainPVZ.PVZWithReceptions, error) {
	args := m.Called(ctx, startDate, endDate, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainPVZ.PVZWithReceptions), args.Error(1)
}

func (m *MockPVZService) GetAll(ctx context.Context) ([]*domainPVZ.PVZ, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainPVZ.PVZ), args.Error(1)
}

func (m *MockPVZService) Update(ctx context.Context, pvz *domainPVZ.PVZ, moderatorID uuid.UUID) error {
	args := m.Called(ctx, pvz, moderatorID)
	return args.Error(0)
}

func (m *MockPVZService) Delete(ctx context.Context, id uuid.UUID, moderatorID uuid.UUID) error {
	args := m.Called(ctx, id, moderatorID)
	return args.Error(0)
}

func (m *MockPVZService) List(ctx context.Context, offset, limit int) ([]*domainPVZ.PVZ, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainPVZ.PVZ), args.Error(1)
}

func TestPVZHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupAuth      func(context.Context) context.Context
		setupMock      func(*MockPVZService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "успешное создание",
			requestBody: map[string]string{
				"city": "Москва",
			},
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserID(ctx, uuid.New())
			},
			setupMock: func(m *MockPVZService) {
				m.On("Create", mock.Anything, "Москва", mock.AnythingOfType("uuid.UUID")).
					Return(&domainPVZ.PVZ{
						ID:        uuid.New(),
						City:      "Москва",
						CreatedAt: time.Now(),
					}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "неверный формат запроса",
			requestBody: map[string]interface{}{
				"city": 123,
			},
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserID(ctx, uuid.New())
			},
			setupMock:      func(m *MockPVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат запроса",
		},
		{
			name: "пустой город",
			requestBody: map[string]string{
				"city": "",
			},
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserID(ctx, uuid.New())
			},
			setupMock:      func(m *MockPVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "город не может быть пустым",
		},
		{
			name: "без авторизации",
			requestBody: map[string]string{
				"city": "Москва",
			},
			setupAuth:      func(ctx context.Context) context.Context { return ctx },
			setupMock:      func(m *MockPVZService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "требуется авторизация",
		},
		{
			name: "ошибка валидации города",
			requestBody: map[string]string{
				"city": "Invalid City",
			},
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserID(ctx, uuid.New())
			},
			setupMock: func(m *MockPVZService) {
				m.On("Create", mock.Anything, "Invalid City", mock.AnythingOfType("uuid.UUID")).
					Return(nil, servicePVZ.ErrInvalidCity)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверное название города",
		},
		{
			name: "ПВЗ уже существует",
			requestBody: map[string]string{
				"city": "Москва",
			},
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserID(ctx, uuid.New())
			},
			setupMock: func(m *MockPVZService) {
				m.On("Create", mock.Anything, "Москва", mock.AnythingOfType("uuid.UUID")).
					Return(nil, servicePVZ.ErrPVZAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   "ПВЗ уже существует",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPVZService)
			tt.setupMock(mockService)

			handler := NewPVZHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(body))
			req = req.WithContext(tt.setupAuth(req.Context()))
			rec := httptest.NewRecorder()

			handler.Create(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response["error"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestPVZHandler_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		pvzID          string
		setupMock      func(*MockPVZService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:  "успешное получение",
			pvzID: uuid.New().String(),
			setupMock: func(m *MockPVZService) {
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&domainPVZ.PVZ{
						ID:        uuid.New(),
						City:      "Москва",
						CreatedAt: time.Now(),
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "неверный формат ID",
			pvzID:          "invalid-uuid",
			setupMock:      func(m *MockPVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат ID ПВЗ",
		},
		{
			name:  "ПВЗ не найден",
			pvzID: uuid.New().String(),
			setupMock: func(m *MockPVZService) {
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, servicePVZ.ErrPVZNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "ПВЗ не найден",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPVZService)
			tt.setupMock(mockService)

			handler := NewPVZHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/pvz/"+tt.pvzID, nil)
			rec := httptest.NewRecorder()

			// Создаем роутер Chi для тестирования URL-параметров
			r := chi.NewRouter()
			r.Get("/pvz/{id}", handler.GetByID)
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response["error"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestPVZHandler_GetWithReceptions(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		setupMock      func(*MockPVZService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "успешное получение",
			queryParams: map[string]string{
				"start_date": time.Now().Format(time.RFC3339),
				"end_date":   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"page":       "1",
				"limit":      "10",
			},
			setupMock: func(m *MockPVZService) {
				m.On("GetWithReceptions",
					mock.Anything,
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("time.Time"),
					1, 10,
				).Return([]*domainPVZ.PVZWithReceptions{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "неверный формат даты начала",
			queryParams: map[string]string{
				"start_date": "invalid-date",
				"end_date":   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			},
			setupMock:      func(m *MockPVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат даты начала",
		},
		{
			name: "неверный формат даты окончания",
			queryParams: map[string]string{
				"start_date": time.Now().Format(time.RFC3339),
				"end_date":   "invalid-date",
			},
			setupMock:      func(m *MockPVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат даты окончания",
		},
		{
			name: "неверный формат страницы",
			queryParams: map[string]string{
				"start_date": time.Now().Format(time.RFC3339),
				"end_date":   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"page":       "invalid",
			},
			setupMock:      func(m *MockPVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат страницы",
		},
		{
			name: "неверный формат лимита",
			queryParams: map[string]string{
				"start_date": time.Now().Format(time.RFC3339),
				"end_date":   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"limit":      "invalid",
			},
			setupMock:      func(m *MockPVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат лимита",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPVZService)
			tt.setupMock(mockService)

			handler := NewPVZHandler(mockService)

			// Создаем URL с query-параметрами
			req := httptest.NewRequest(http.MethodGet, "/pvz", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			rec := httptest.NewRecorder()

			handler.GetWithReceptions(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response["error"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestPVZHandler_UpdatePVZ(t *testing.T) {
	tests := []struct {
		name           string
		pvzID          string
		requestBody    map[string]interface{}
		moderatorID    uuid.UUID
		mockSetup      func(*MockPVZService)
		expectedStatus int
	}{
		{
			name:  "успешное обновление ПВЗ",
			pvzID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"city": "Санкт-Петербург",
			},
			moderatorID: uuid.New(),
			mockSetup: func(m *MockPVZService) {
				m.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "неверный формат ID",
			pvzID:       "invalid-uuid",
			requestBody: map[string]interface{}{},
			moderatorID: uuid.New(),
			mockSetup: func(m *MockPVZService) {
				// Для неверного ID мок не нужен, так как до вызова сервиса не дойдет
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "ПВЗ не найден",
			pvzID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"city": "Москва",
			},
			moderatorID: uuid.New(),
			mockSetup: func(m *MockPVZService) {
				m.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(servicePVZ.ErrPVZNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:  "отсутствие авторизации",
			pvzID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"city": "Москва",
			},
			moderatorID: uuid.UUID{}, // Пустой ID для случая без авторизации
			mockSetup: func(m *MockPVZService) {
				// Мок не нужен, так как до вызова сервиса не дойдет
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPVZService)
			tt.mockSetup(mockService)

			handler := NewPVZHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/pvz/"+tt.pvzID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Устанавливаем контекст с ID пользователя
			ctx := context.Background()
			if tt.moderatorID != uuid.Nil {
				ctx = context.WithValue(ctx, middleware.UserIDKey, tt.moderatorID.String())
			}
			req = req.WithContext(ctx)

			// Добавляем параметры маршрутизации
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("id", tt.pvzID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			rr := httptest.NewRecorder()

			handler.UpdatePVZ(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestPVZHandler_DeletePVZ(t *testing.T) {
	tests := []struct {
		name           string
		pvzID          string
		moderatorID    uuid.UUID
		mockSetup      func(*MockPVZService)
		expectedStatus int
	}{
		{
			name:        "успешное удаление ПВЗ",
			pvzID:       uuid.New().String(),
			moderatorID: uuid.New(),
			mockSetup: func(m *MockPVZService) {
				m.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "неверный формат ID",
			pvzID:       "invalid-uuid",
			moderatorID: uuid.New(),
			mockSetup: func(m *MockPVZService) {
				// Для неверного ID мок не нужен, так как до вызова сервиса не дойдет
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "ПВЗ не найден",
			pvzID:       uuid.New().String(),
			moderatorID: uuid.New(),
			mockSetup: func(m *MockPVZService) {
				m.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(servicePVZ.ErrPVZNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "отсутствие авторизации",
			pvzID:       uuid.New().String(),
			moderatorID: uuid.UUID{}, // Пустой ID для случая без авторизации
			mockSetup: func(m *MockPVZService) {
				// Мок не нужен, так как до вызова сервиса не дойдет
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPVZService)
			tt.mockSetup(mockService)

			handler := NewPVZHandler(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/pvz/"+tt.pvzID, nil)

			// Устанавливаем контекст с ID пользователя
			ctx := context.Background()
			if tt.moderatorID != uuid.Nil {
				ctx = context.WithValue(ctx, middleware.UserIDKey, tt.moderatorID.String())
			}
			req = req.WithContext(ctx)

			// Добавляем параметры маршрутизации
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("id", tt.pvzID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			rr := httptest.NewRecorder()

			handler.DeletePVZ(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestPVZHandler_ListPVZ(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockPVZService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "успешное получение списка ПВЗ",
			queryParams: "?offset=0&limit=10",
			mockSetup: func(m *MockPVZService) {
				pvzList := []*domainPVZ.PVZ{
					{
						ID:        uuid.New(),
						CreatedAt: time.Now().UTC(),
						City:      "Москва",
					},
					{
						ID:        uuid.New(),
						CreatedAt: time.Now().UTC(),
						City:      "Санкт-Петербург",
					},
				}
				m.On("List", mock.Anything, 0, 10).Return(pvzList, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []map[string]interface{}{
				{
					"id":         mock.Anything,
					"created_at": mock.Anything,
					"city":       "Москва",
				},
				{
					"id":         mock.Anything,
					"created_at": mock.Anything,
					"city":       "Санкт-Петербург",
				},
			},
		},
		{
			name:        "получение списка с неверными параметрами",
			queryParams: "?offset=invalid&limit=invalid",
			mockSetup: func(m *MockPVZService) {
				m.On("List", mock.Anything, 0, 10).Return([]*domainPVZ.PVZ{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []map[string]interface{}{},
		},
		{
			name:        "ошибка сервиса",
			queryParams: "?offset=0&limit=10",
			mockSetup: func(m *MockPVZService) {
				m.On("List", mock.Anything, 0, 10).Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPVZService)
			tt.mockSetup(mockService)

			handler := NewPVZHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/pvz"+tt.queryParams, nil)
			rr := httptest.NewRecorder()

			handler.ListPVZ(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response []map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, len(tt.expectedBody.([]map[string]interface{})), len(response))
				for i, expected := range tt.expectedBody.([]map[string]interface{}) {
					for key, value := range expected {
						if value == mock.Anything {
							assert.NotEmpty(t, response[i][key])
						} else {
							assert.Equal(t, value, response[i][key])
						}
					}
				}
			} else {
				assert.Contains(t, rr.Body.String(), tt.expectedBody.(string))
			}

			mockService.AssertExpectations(t)
		})
	}
}
