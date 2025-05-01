package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/avito/pvz/internal/domain/reception"
	receptionService "github.com/avito/pvz/internal/service/reception"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockReceptionService struct {
	mock.Mock
}

func (m *mockReceptionService) Create(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception.Reception), args.Error(1)
}

func (m *mockReceptionService) GetByID(ctx context.Context, id uuid.UUID) (*reception.Reception, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception.Reception), args.Error(1)
}

func (m *mockReceptionService) Close(ctx context.Context, pvzID uuid.UUID) error {
	args := m.Called(ctx, pvzID)
	return args.Error(0)
}

func (m *mockReceptionService) GetOpenByPVZID(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception.Reception), args.Error(1)
}

func (m *mockReceptionService) List(ctx context.Context, offset, limit int) ([]*reception.Reception, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*reception.Reception), args.Error(1)
}

func TestReceptionHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mockReceptionService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "успешное создание",
			requestBody: map[string]interface{}{
				"pvz_id": uuid.New().String(),
			},
			setupMocks: func(rs *mockReceptionService) {
				pvzID := uuid.New()
				rs.On("Create", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&reception.Reception{
						ID:       uuid.New(),
						PVZID:    pvzID,
						DateTime: time.Now(),
						Status:   reception.StatusInProgress,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "неверный формат запроса",
			requestBody: map[string]interface{}{
				"pvz_id": "invalid-uuid",
			},
			setupMocks:     func(rs *mockReceptionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат запроса",
		},
		{
			name: "ПВЗ не найден",
			requestBody: map[string]interface{}{
				"pvz_id": uuid.New().String(),
			},
			setupMocks: func(rs *mockReceptionService) {
				rs.On("Create", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, receptionService.ErrPVZNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "ПВЗ не найден",
		},
		{
			name: "приемка уже открыта",
			requestBody: map[string]interface{}{
				"pvz_id": uuid.New().String(),
			},
			setupMocks: func(rs *mockReceptionService) {
				rs.On("Create", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, receptionService.ErrReceptionAlreadyOpen)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "у ПВЗ уже есть открытая приемка",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := new(mockReceptionService)
			tt.setupMocks(service)

			handler := NewReceptionHandler(service)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/reception", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.Create(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response["error"])
			}

			service.AssertExpectations(t)
		})
	}
}

func TestReceptionHandler_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		receptionID    string
		setupMocks     func(*mockReceptionService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "успешное получение",
			receptionID: uuid.New().String(),
			setupMocks: func(rs *mockReceptionService) {
				receptionID := uuid.New()
				rs.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&reception.Reception{
						ID:       receptionID,
						PVZID:    uuid.New(),
						DateTime: time.Now(),
						Status:   reception.StatusInProgress,
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "неверный формат ID",
			receptionID:    "invalid-uuid",
			setupMocks:     func(rs *mockReceptionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат ID приемки",
		},
		{
			name:        "приемка не найдена",
			receptionID: uuid.New().String(),
			setupMocks: func(rs *mockReceptionService) {
				rs.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, receptionService.ErrReceptionNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "приемка не найдена",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := new(mockReceptionService)
			tt.setupMocks(service)

			handler := NewReceptionHandler(service)

			req := httptest.NewRequest(http.MethodGet, "/reception/"+tt.receptionID, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chi.NewRouteContext()))
			chi.RouteContext(req.Context()).URLParams.Add("id", tt.receptionID)

			rec := httptest.NewRecorder()

			handler.GetByID(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response["error"])
			}

			service.AssertExpectations(t)
		})
	}
}

func TestReceptionHandler_Close(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mockReceptionService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "успешное закрытие",
			requestBody: map[string]interface{}{
				"pvz_id": uuid.New().String(),
			},
			setupMocks: func(rs *mockReceptionService) {
				rs.On("Close", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "неверный формат запроса",
			requestBody: map[string]interface{}{
				"pvz_id": "invalid-uuid",
			},
			setupMocks:     func(rs *mockReceptionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат ID ПВЗ",
		},
		{
			name: "ПВЗ не найден",
			requestBody: map[string]interface{}{
				"pvz_id": uuid.New().String(),
			},
			setupMocks: func(rs *mockReceptionService) {
				rs.On("Close", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(receptionService.ErrPVZNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "ПВЗ не найден",
		},
		{
			name: "приемка не найдена",
			requestBody: map[string]interface{}{
				"pvz_id": uuid.New().String(),
			},
			setupMocks: func(rs *mockReceptionService) {
				rs.On("Close", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(receptionService.ErrReceptionNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "приемка не найдена",
		},
		{
			name: "приемка уже закрыта",
			requestBody: map[string]interface{}{
				"pvz_id": uuid.New().String(),
			},
			setupMocks: func(rs *mockReceptionService) {
				rs.On("Close", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(receptionService.ErrReceptionAlreadyClose)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "приемка уже закрыта",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := new(mockReceptionService)
			tt.setupMocks(service)

			handler := NewReceptionHandler(service)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/reception/close", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.Close(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response["error"])
			}

			service.AssertExpectations(t)
		})
	}
}

func TestReceptionHandler_ListReceptions(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		setupMocks     func(*mockReceptionService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "успешное получение списка",
			queryParams: map[string]string{
				"offset": "0",
				"limit":  "10",
			},
			setupMocks: func(rs *mockReceptionService) {
				rs.On("List", mock.Anything, 0, 10).
					Return([]*reception.Reception{
						{
							ID:       uuid.New(),
							PVZID:    uuid.New(),
							DateTime: time.Now(),
							Status:   reception.StatusInProgress,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "без параметров пагинации",
			setupMocks: func(rs *mockReceptionService) {
				rs.On("List", mock.Anything, 0, 10).
					Return([]*reception.Reception{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "ошибка сервиса",
			queryParams: map[string]string{
				"offset": "0",
				"limit":  "10",
			},
			setupMocks: func(rs *mockReceptionService) {
				rs.On("List", mock.Anything, 0, 10).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := new(mockReceptionService)
			tt.setupMocks(service)

			handler := NewReceptionHandler(service)

			req := httptest.NewRequest(http.MethodGet, "/reception", nil)
			if tt.queryParams != nil {
				q := req.URL.Query()
				for key, value := range tt.queryParams {
					q.Add(key, value)
				}
				req.URL.RawQuery = q.Encode()
			}

			rec := httptest.NewRecorder()

			handler.ListReceptions(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response["error"])
			}

			service.AssertExpectations(t)
		})
	}
}
