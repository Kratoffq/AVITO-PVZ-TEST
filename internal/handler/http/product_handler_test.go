package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/avito/pvz/internal/domain/user"
	productService "github.com/avito/pvz/internal/service/product"
	"github.com/avito/pvz/pkg/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockProductRepo struct {
	mock.Mock
}

func (m *mockProductRepo) Create(ctx context.Context, product *product.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *mockProductRepo) CreateBatch(ctx context.Context, products []*product.Product) error {
	args := m.Called(ctx, products)
	return args.Error(0)
}

func (m *mockProductRepo) DeleteLast(ctx context.Context, receptionID uuid.UUID) error {
	args := m.Called(ctx, receptionID)
	return args.Error(0)
}

func (m *mockProductRepo) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

func (m *mockProductRepo) GetByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]*product.Product, error) {
	args := m.Called(ctx, receptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*product.Product), args.Error(1)
}

func (m *mockProductRepo) List(ctx context.Context, offset, limit int) ([]*product.Product, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*product.Product), args.Error(1)
}

func (m *mockProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockProductRepo) GetLast(ctx context.Context, receptionID uuid.UUID) (*product.Product, error) {
	args := m.Called(ctx, receptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

func (m *mockProductRepo) Update(ctx context.Context, product *product.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

type mockReceptionRepo struct {
	mock.Mock
}

func (m *mockReceptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*reception.Reception, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception.Reception), args.Error(1)
}

func (m *mockReceptionRepo) Create(ctx context.Context, reception *reception.Reception) error {
	args := m.Called(ctx, reception)
	return args.Error(0)
}

func (m *mockReceptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockReceptionRepo) GetLastOpen(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception.Reception), args.Error(1)
}

func (m *mockReceptionRepo) GetOpenByPVZID(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception.Reception), args.Error(1)
}

func (m *mockReceptionRepo) GetProducts(ctx context.Context, receptionID uuid.UUID) ([]*product.Product, error) {
	args := m.Called(ctx, receptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*product.Product), args.Error(1)
}

func (m *mockReceptionRepo) List(ctx context.Context, offset, limit int) ([]*reception.Reception, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*reception.Reception), args.Error(1)
}

func (m *mockReceptionRepo) Update(ctx context.Context, reception *reception.Reception) error {
	args := m.Called(ctx, reception)
	return args.Error(0)
}

type mockTxManager struct {
	mock.Mock
}

func (m *mockTxManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func TestProductHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupAuth      func(context.Context) context.Context
		setupMocks     func(*mockProductRepo, *mockReceptionRepo, *mockTxManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "успешное создание",
			requestBody: map[string]interface{}{
				"reception_id": uuid.New().String(),
				"type":         "electronics",
			},
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserRole(ctx, user.RoleEmployee)
			},
			setupMocks: func(pr *mockProductRepo, rr *mockReceptionRepo, tm *mockTxManager) {
				receptionID := uuid.New()
				rr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&reception.Reception{
						ID:     receptionID,
						Status: reception.StatusInProgress,
					}, nil)

				pr.On("Create", mock.Anything, mock.AnythingOfType("*product.Product")).
					Return(nil).Maybe()

				tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(nil).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(args.Get(0).(context.Context))
				})
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "неверный формат запроса",
			requestBody: map[string]interface{}{
				"reception_id": "invalid-uuid",
				"type":         "invalid-type",
			},
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserRole(ctx, user.RoleEmployee)
			},
			setupMocks:     func(pr *mockProductRepo, rr *mockReceptionRepo, tm *mockTxManager) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат ID приемки",
		},
		{
			name: "приемка не найдена",
			requestBody: map[string]interface{}{
				"reception_id": uuid.New().String(),
				"type":         "electronics",
			},
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserRole(ctx, user.RoleEmployee)
			},
			setupMocks: func(pr *mockProductRepo, rr *mockReceptionRepo, tm *mockTxManager) {
				rr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, productService.ErrReceptionNotFound).Maybe()

				tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(productService.ErrReceptionNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "приемка не найдена",
		},
		{
			name: "приемка уже закрыта",
			requestBody: map[string]interface{}{
				"reception_id": uuid.New().String(),
				"type":         "electronics",
			},
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserRole(ctx, user.RoleEmployee)
			},
			setupMocks: func(pr *mockProductRepo, rr *mockReceptionRepo, tm *mockTxManager) {
				receptionID := uuid.New()
				rr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&reception.Reception{
						ID:     receptionID,
						Status: reception.StatusClose,
					}, nil).Maybe()

				tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(productService.ErrReceptionAlreadyClose)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "приемка уже закрыта",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(mockProductRepo)
			receptionRepo := new(mockReceptionRepo)
			txManager := new(mockTxManager)
			tt.setupMocks(productRepo, receptionRepo, txManager)

			service := productService.New(productRepo, receptionRepo, txManager)
			handler := NewProductHandler(service)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/product", bytes.NewReader(body))
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

			productRepo.AssertExpectations(t)
			receptionRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}

func TestProductHandler_CreateBatch(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupAuth      func(context.Context) context.Context
		setupMocks     func(*mockProductRepo, *mockReceptionRepo, *mockTxManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "успешное создание",
			requestBody: map[string]interface{}{
				"reception_id": uuid.New().String(),
				"types":        []string{"electronics", "clothing"},
			},
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserRole(ctx, user.RoleEmployee)
			},
			setupMocks: func(pr *mockProductRepo, rr *mockReceptionRepo, tm *mockTxManager) {
				receptionID := uuid.New()
				rr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&reception.Reception{
						ID:     receptionID,
						Status: reception.StatusInProgress,
					}, nil)

				pr.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*product.Product")).
					Return(nil).Maybe()

				tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(nil).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(args.Get(0).(context.Context))
				})
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "неверный формат запроса",
			requestBody: map[string]interface{}{
				"reception_id": "invalid-uuid",
				"types":        []string{"invalid-type"},
			},
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserRole(ctx, user.RoleEmployee)
			},
			setupMocks:     func(pr *mockProductRepo, rr *mockReceptionRepo, tm *mockTxManager) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат ID приемки",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(mockProductRepo)
			receptionRepo := new(mockReceptionRepo)
			txManager := new(mockTxManager)
			tt.setupMocks(productRepo, receptionRepo, txManager)

			service := productService.New(productRepo, receptionRepo, txManager)
			handler := NewProductHandler(service)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/product/batch", bytes.NewReader(body))
			req = req.WithContext(tt.setupAuth(req.Context()))
			rec := httptest.NewRecorder()

			handler.CreateBatch(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response["error"])
			}

			productRepo.AssertExpectations(t)
			receptionRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}

func TestProductHandler_DeleteLast(t *testing.T) {
	tests := []struct {
		name           string
		receptionID    string
		setupAuth      func(context.Context) context.Context
		setupMocks     func(*mockProductRepo, *mockReceptionRepo, *mockTxManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "успешное удаление",
			receptionID: uuid.New().String(),
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserRole(ctx, user.RoleEmployee)
			},
			setupMocks: func(pr *mockProductRepo, rr *mockReceptionRepo, tm *mockTxManager) {
				receptionID := uuid.New()
				rr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&reception.Reception{
						ID:     receptionID,
						Status: reception.StatusInProgress,
					}, nil)

				pr.On("DeleteLast", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil).Maybe()

				tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(nil).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(args.Get(0).(context.Context))
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "неверный формат ID",
			receptionID: "invalid-uuid",
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserRole(ctx, user.RoleEmployee)
			},
			setupMocks:     func(pr *mockProductRepo, rr *mockReceptionRepo, tm *mockTxManager) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат ID приемки",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(mockProductRepo)
			receptionRepo := new(mockReceptionRepo)
			txManager := new(mockTxManager)
			tt.setupMocks(productRepo, receptionRepo, txManager)

			service := productService.New(productRepo, receptionRepo, txManager)
			handler := NewProductHandler(service)

			req := httptest.NewRequest(http.MethodDelete, "/product/last/"+tt.receptionID, nil)
			req = req.WithContext(tt.setupAuth(req.Context()))

			r := chi.NewRouter()
			r.Delete("/product/last/{reception_id}", handler.DeleteLast)

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response["error"])
			}

			productRepo.AssertExpectations(t)
			receptionRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}

func TestProductHandler_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		productID      string
		setupAuth      func(context.Context) context.Context
		setupMocks     func(*mockProductRepo, *mockReceptionRepo, *mockTxManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "успешное получение",
			productID: uuid.New().String(),
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserRole(ctx, user.RoleEmployee)
			},
			setupMocks: func(pr *mockProductRepo, rr *mockReceptionRepo, tm *mockTxManager) {
				pr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&product.Product{
						ID:          uuid.New(),
						ReceptionID: uuid.New(),
						Type:        "electronics",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "неверный формат ID",
			productID: "invalid-uuid",
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserRole(ctx, user.RoleEmployee)
			},
			setupMocks:     func(pr *mockProductRepo, rr *mockReceptionRepo, tm *mockTxManager) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат ID товара",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(mockProductRepo)
			receptionRepo := new(mockReceptionRepo)
			txManager := new(mockTxManager)
			tt.setupMocks(productRepo, receptionRepo, txManager)

			service := productService.New(productRepo, receptionRepo, txManager)
			handler := NewProductHandler(service)

			req := httptest.NewRequest(http.MethodGet, "/product/"+tt.productID, nil)
			req = req.WithContext(tt.setupAuth(req.Context()))

			r := chi.NewRouter()
			r.Get("/product/{id}", handler.GetByID)

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response["error"])
			}

			productRepo.AssertExpectations(t)
			receptionRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}

func TestProductHandler_GetByReceptionID(t *testing.T) {
	tests := []struct {
		name           string
		receptionID    string
		setupAuth      func(context.Context) context.Context
		setupMocks     func(*mockProductRepo, *mockReceptionRepo, *mockTxManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "успешное получение",
			receptionID: uuid.New().String(),
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserRole(ctx, user.RoleEmployee)
			},
			setupMocks: func(pr *mockProductRepo, rr *mockReceptionRepo, tm *mockTxManager) {
				pr.On("GetByReceptionID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return([]*product.Product{
						{
							ID:          uuid.New(),
							ReceptionID: uuid.New(),
							Type:        "electronics",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "неверный формат ID",
			receptionID: "invalid-uuid",
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserRole(ctx, user.RoleEmployee)
			},
			setupMocks:     func(pr *mockProductRepo, rr *mockReceptionRepo, tm *mockTxManager) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат ID приемки",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(mockProductRepo)
			receptionRepo := new(mockReceptionRepo)
			txManager := new(mockTxManager)
			tt.setupMocks(productRepo, receptionRepo, txManager)

			service := productService.New(productRepo, receptionRepo, txManager)
			handler := NewProductHandler(service)

			req := httptest.NewRequest(http.MethodGet, "/product/reception/"+tt.receptionID, nil)
			req = req.WithContext(tt.setupAuth(req.Context()))

			r := chi.NewRouter()
			r.Get("/product/reception/{reception_id}", handler.GetByReceptionID)

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response["error"])
			}

			productRepo.AssertExpectations(t)
			receptionRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}

func TestProductHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		setupAuth      func(context.Context) context.Context
		setupMocks     func(*mockProductRepo, *mockReceptionRepo, *mockTxManager)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "успешное получение списка",
			queryParams: map[string]string{
				"offset": "0",
				"limit":  "10",
			},
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserRole(ctx, user.RoleEmployee)
			},
			setupMocks: func(pr *mockProductRepo, rr *mockReceptionRepo, tm *mockTxManager) {
				pr.On("List", mock.Anything, 0, 10).
					Return([]*product.Product{
						{
							ID:          uuid.New(),
							ReceptionID: uuid.New(),
							Type:        "electronics",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "без параметров пагинации",
			setupAuth: func(ctx context.Context) context.Context {
				return auth.WithUserRole(ctx, user.RoleEmployee)
			},
			setupMocks: func(pr *mockProductRepo, rr *mockReceptionRepo, tm *mockTxManager) {
				pr.On("List", mock.Anything, 0, 10).
					Return([]*product.Product{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productRepo := new(mockProductRepo)
			receptionRepo := new(mockReceptionRepo)
			txManager := new(mockTxManager)
			tt.setupMocks(productRepo, receptionRepo, txManager)

			service := productService.New(productRepo, receptionRepo, txManager)
			handler := NewProductHandler(service)

			req := httptest.NewRequest(http.MethodGet, "/product", nil)
			if tt.queryParams != nil {
				q := req.URL.Query()
				for key, value := range tt.queryParams {
					q.Add(key, value)
				}
				req.URL.RawQuery = q.Encode()
			}
			req = req.WithContext(tt.setupAuth(req.Context()))

			rec := httptest.NewRecorder()
			handler.List(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				var response map[string]string
				err := json.NewDecoder(rec.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response["error"])
			}

			productRepo.AssertExpectations(t)
			receptionRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}
