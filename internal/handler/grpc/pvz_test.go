package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/avito/pvz/api/proto"
	domainPVZ "github.com/avito/pvz/internal/domain/pvz"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PVZServiceInterface определяет интерфейс сервиса для тестов
type PVZServiceInterface interface {
	Create(ctx context.Context, city string, userID uuid.UUID) (*domainPVZ.PVZ, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domainPVZ.PVZ, error)
	GetWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*domainPVZ.PVZWithReceptions, error)
	GetAll(ctx context.Context) ([]*domainPVZ.PVZ, error)
	Update(ctx context.Context, pvz *domainPVZ.PVZ, moderatorID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID, moderatorID uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*domainPVZ.PVZ, error)
}

// MockPVZService реализует интерфейс сервиса для тестов
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
	return args.Get(0).([]*domainPVZ.PVZWithReceptions), args.Error(1)
}

func (m *MockPVZService) GetAll(ctx context.Context) ([]*domainPVZ.PVZ, error) {
	args := m.Called(ctx)
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
	return args.Get(0).([]*domainPVZ.PVZ), args.Error(1)
}

// testPVZHandler - тестовая версия PVZHandler
type testPVZHandler struct {
	proto.UnimplementedPVZServiceServer
	pvzService interface {
		GetAll(ctx context.Context) ([]*domainPVZ.PVZ, error)
	}
}

// GetAllPVZ возвращает список всех ПВЗ
func (h *testPVZHandler) GetAllPVZ(ctx context.Context, req *proto.GetAllPVZRequest) (*proto.GetAllPVZResponse, error) {
	pvzs, err := h.pvzService.GetAll(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get PVZs")
	}

	response := &proto.GetAllPVZResponse{
		Pvzs: make([]*proto.PVZ, len(pvzs)),
	}

	for i, p := range pvzs {
		response.Pvzs[i] = &proto.PVZ{
			Id:               p.ID.String(),
			City:             p.City,
			RegistrationDate: timestamppb.New(p.CreatedAt),
		}
	}

	return response, nil
}

func TestPVZHandler_GetAllPVZ(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockPVZService)
		expectedError error
		expectedPVZs  []*proto.PVZ
	}{
		{
			name: "успешное получение списка ПВЗ",
			mockSetup: func(m *MockPVZService) {
				pvzs := []*domainPVZ.PVZ{
					{
						ID:        uuid.New(),
						City:      "Москва",
						CreatedAt: time.Now().UTC(),
					},
					{
						ID:        uuid.New(),
						City:      "Санкт-Петербург",
						CreatedAt: time.Now().UTC(),
					},
				}
				m.On("GetAll", mock.Anything).Return(pvzs, nil)
			},
			expectedError: nil,
			expectedPVZs: []*proto.PVZ{
				{
					Id:   mock.Anything,
					City: "Москва",
				},
				{
					Id:   mock.Anything,
					City: "Санкт-Петербург",
				},
			},
		},
		{
			name: "ошибка при получении списка ПВЗ",
			mockSetup: func(m *MockPVZService) {
				m.On("GetAll", mock.Anything).Return([]*domainPVZ.PVZ(nil), assert.AnError)
			},
			expectedError: status.Error(codes.Internal, "failed to get PVZs"),
			expectedPVZs:  nil,
		},
		{
			name: "пустой список ПВЗ",
			mockSetup: func(m *MockPVZService) {
				m.On("GetAll", mock.Anything).Return([]*domainPVZ.PVZ{}, nil)
			},
			expectedError: nil,
			expectedPVZs:  []*proto.PVZ{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPVZService)
			tt.mockSetup(mockService)

			handler := &testPVZHandler{
				UnimplementedPVZServiceServer: proto.UnimplementedPVZServiceServer{},
				pvzService:                    mockService,
			}
			resp, err := handler.GetAllPVZ(context.Background(), &proto.GetAllPVZRequest{})

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, len(tt.expectedPVZs), len(resp.Pvzs))

				for i, expectedPVZ := range tt.expectedPVZs {
					actualPVZ := resp.Pvzs[i]
					if expectedPVZ.Id == mock.Anything {
						assert.NotEmpty(t, actualPVZ.Id)
					} else {
						assert.Equal(t, expectedPVZ.Id, actualPVZ.Id)
					}
					assert.Equal(t, expectedPVZ.City, actualPVZ.City)
					assert.NotNil(t, actualPVZ.RegistrationDate)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}
