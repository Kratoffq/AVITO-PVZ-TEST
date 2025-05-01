package service_test

import (
	"context"
	"testing"
	"time"

	domainPVZ "github.com/avito/pvz/internal/domain/pvz"
	domainUser "github.com/avito/pvz/internal/domain/user"
	"github.com/avito/pvz/internal/service/pvz"
	"github.com/avito/pvz/test/unit/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuditLog реализует интерфейс audit.AuditLog
type MockAuditLog struct {
	mock.Mock
}

func (m *MockAuditLog) LogPVZCreation(ctx context.Context, pvzID, userID uuid.UUID) error {
	args := m.Called(ctx, pvzID, userID)
	return args.Error(0)
}

func TestPVZService_Create(t *testing.T) {
	tests := []struct {
		name    string
		city    string
		mock    func(*mocks.MockPVZRepository, *mocks.MockUserRepository, *mocks.MockTransactionManager, *MockAuditLog)
		wantErr bool
	}{
		{
			name: "successful creation",
			city: "Moscow",
			mock: func(pvzRepo *mocks.MockPVZRepository, userRepo *mocks.MockUserRepository, txManager *mocks.MockTransactionManager, auditLog *MockAuditLog) {
				pvzRepo.On("Create", mock.Anything, mock.AnythingOfType("*domainPVZ.PVZ")).Return(nil)
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domainUser.User{Role: domainUser.RoleAdmin}, nil)
				txManager.On("WithinTransaction", mock.Anything, mock.Anything).Return(nil)
				auditLog.On("LogPVZCreation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "empty city",
			city: "",
			mock: func(pvzRepo *mocks.MockPVZRepository, userRepo *mocks.MockUserRepository, txManager *mocks.MockTransactionManager, auditLog *MockAuditLog) {
				txManager.On("WithinTransaction", mock.Anything, mock.Anything).Return(pvz.ErrInvalidCity)
			},
			wantErr: true,
		},
		{
			name: "repository error",
			city: "Moscow",
			mock: func(pvzRepo *mocks.MockPVZRepository, userRepo *mocks.MockUserRepository, txManager *mocks.MockTransactionManager, auditLog *MockAuditLog) {
				pvzRepo.On("Create", mock.Anything, mock.AnythingOfType("*domainPVZ.PVZ")).Return(assert.AnError)
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domainUser.User{Role: domainUser.RoleAdmin}, nil)
				txManager.On("WithinTransaction", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(mocks.MockPVZRepository)
			userRepo := new(mocks.MockUserRepository)
			txManager := new(mocks.MockTransactionManager)
			auditLog := new(MockAuditLog)
			tt.mock(pvzRepo, userRepo, txManager, auditLog)

			defaultUser := &domainUser.User{Role: domainUser.RoleAdmin}
			service := pvz.New(pvzRepo, userRepo, txManager, auditLog, defaultUser)
			_, err := service.Create(context.Background(), tt.city, uuid.New())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPVZService_GetByID(t *testing.T) {
	id := uuid.New()
	expectedPVZ := &domainPVZ.PVZ{
		ID:        id,
		CreatedAt: time.Now(),
		City:      "Moscow",
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		mock    func(*mocks.MockPVZRepository, *mocks.MockUserRepository, *mocks.MockTransactionManager, *MockAuditLog)
		want    *domainPVZ.PVZ
		wantErr bool
	}{
		{
			name: "successful get",
			id:   id,
			mock: func(pvzRepo *mocks.MockPVZRepository, userRepo *mocks.MockUserRepository, txManager *mocks.MockTransactionManager, auditLog *MockAuditLog) {
				pvzRepo.On("GetByID", mock.Anything, id).Return(expectedPVZ, nil)
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domainUser.User{Role: domainUser.RoleAdmin}, nil)
			},
			want:    expectedPVZ,
			wantErr: false,
		},
		{
			name: "not found",
			id:   uuid.New(),
			mock: func(pvzRepo *mocks.MockPVZRepository, userRepo *mocks.MockUserRepository, txManager *mocks.MockTransactionManager, auditLog *MockAuditLog) {
				pvzRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, domainPVZ.ErrNotFound)
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domainUser.User{Role: domainUser.RoleAdmin}, nil)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(mocks.MockPVZRepository)
			userRepo := new(mocks.MockUserRepository)
			txManager := new(mocks.MockTransactionManager)
			auditLog := new(MockAuditLog)
			tt.mock(pvzRepo, userRepo, txManager, auditLog)

			defaultUser := &domainUser.User{Role: domainUser.RoleAdmin}
			service := pvz.New(pvzRepo, userRepo, txManager, auditLog, defaultUser)
			got, err := service.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestPVZService_Update(t *testing.T) {
	id := uuid.New()
	p := &domainPVZ.PVZ{
		ID:        id,
		CreatedAt: time.Now(),
		City:      "Moscow",
	}

	tests := []struct {
		name    string
		pvz     *domainPVZ.PVZ
		mock    func(*mocks.MockPVZRepository, *mocks.MockUserRepository, *mocks.MockTransactionManager, *MockAuditLog)
		wantErr bool
	}{
		{
			name: "successful update",
			pvz:  p,
			mock: func(pvzRepo *mocks.MockPVZRepository, userRepo *mocks.MockUserRepository, txManager *mocks.MockTransactionManager, auditLog *MockAuditLog) {
				pvzRepo.On("Update", mock.Anything, p).Return(nil)
				pvzRepo.On("GetByID", mock.Anything, p.ID).Return(p, nil)
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domainUser.User{Role: domainUser.RoleAdmin}, nil)
				txManager.On("WithinTransaction", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			pvz:  p,
			mock: func(pvzRepo *mocks.MockPVZRepository, userRepo *mocks.MockUserRepository, txManager *mocks.MockTransactionManager, auditLog *MockAuditLog) {
				pvzRepo.On("GetByID", mock.Anything, p.ID).Return(nil, domainPVZ.ErrNotFound)
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domainUser.User{Role: domainUser.RoleAdmin}, nil)
				txManager.On("WithinTransaction", mock.Anything, mock.Anything).Return(domainPVZ.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(mocks.MockPVZRepository)
			userRepo := new(mocks.MockUserRepository)
			txManager := new(mocks.MockTransactionManager)
			auditLog := new(MockAuditLog)
			tt.mock(pvzRepo, userRepo, txManager, auditLog)

			defaultUser := &domainUser.User{Role: domainUser.RoleAdmin}
			service := pvz.New(pvzRepo, userRepo, txManager, auditLog, defaultUser)
			err := service.Update(context.Background(), tt.pvz, uuid.New())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPVZService_Delete(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		mock    func(*mocks.MockPVZRepository, *mocks.MockUserRepository, *mocks.MockTransactionManager, *MockAuditLog)
		wantErr bool
	}{
		{
			name: "successful delete",
			id:   id,
			mock: func(pvzRepo *mocks.MockPVZRepository, userRepo *mocks.MockUserRepository, txManager *mocks.MockTransactionManager, auditLog *MockAuditLog) {
				pvzRepo.On("Delete", mock.Anything, id).Return(nil)
				pvzRepo.On("GetByID", mock.Anything, id).Return(&domainPVZ.PVZ{ID: id}, nil)
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domainUser.User{Role: domainUser.RoleAdmin}, nil)
				txManager.On("WithinTransaction", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   id,
			mock: func(pvzRepo *mocks.MockPVZRepository, userRepo *mocks.MockUserRepository, txManager *mocks.MockTransactionManager, auditLog *MockAuditLog) {
				pvzRepo.On("GetByID", mock.Anything, id).Return(nil, domainPVZ.ErrNotFound)
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domainUser.User{Role: domainUser.RoleAdmin}, nil)
				txManager.On("WithinTransaction", mock.Anything, mock.Anything).Return(domainPVZ.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(mocks.MockPVZRepository)
			userRepo := new(mocks.MockUserRepository)
			txManager := new(mocks.MockTransactionManager)
			auditLog := new(MockAuditLog)
			tt.mock(pvzRepo, userRepo, txManager, auditLog)

			defaultUser := &domainUser.User{Role: domainUser.RoleAdmin}
			service := pvz.New(pvzRepo, userRepo, txManager, auditLog, defaultUser)
			err := service.Delete(context.Background(), tt.id, uuid.New())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPVZService_List(t *testing.T) {
	expectedPVZs := []*domainPVZ.PVZ{
		{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			City:      "Moscow",
		},
		{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			City:      "St. Petersburg",
		},
	}

	tests := []struct {
		name    string
		offset  int
		limit   int
		mock    func(*mocks.MockPVZRepository, *mocks.MockUserRepository, *mocks.MockTransactionManager, *MockAuditLog)
		want    []*domainPVZ.PVZ
		wantErr bool
	}{
		{
			name:   "successful list",
			offset: 0,
			limit:  10,
			mock: func(pvzRepo *mocks.MockPVZRepository, userRepo *mocks.MockUserRepository, txManager *mocks.MockTransactionManager, auditLog *MockAuditLog) {
				pvzRepo.On("List", mock.Anything, 0, 10).Return(expectedPVZs, nil)
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domainUser.User{Role: domainUser.RoleAdmin}, nil)
			},
			want:    expectedPVZs,
			wantErr: false,
		},
		{
			name:   "repository error",
			offset: 0,
			limit:  10,
			mock: func(pvzRepo *mocks.MockPVZRepository, userRepo *mocks.MockUserRepository, txManager *mocks.MockTransactionManager, auditLog *MockAuditLog) {
				pvzRepo.On("List", mock.Anything, 0, 10).Return([]*domainPVZ.PVZ(nil), assert.AnError)
				userRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domainUser.User{Role: domainUser.RoleAdmin}, nil)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvzRepo := new(mocks.MockPVZRepository)
			userRepo := new(mocks.MockUserRepository)
			txManager := new(mocks.MockTransactionManager)
			auditLog := new(MockAuditLog)
			tt.mock(pvzRepo, userRepo, txManager, auditLog)

			defaultUser := &domainUser.User{Role: domainUser.RoleAdmin}
			service := pvz.New(pvzRepo, userRepo, txManager, auditLog, defaultUser)
			got, err := service.List(context.Background(), tt.offset, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
