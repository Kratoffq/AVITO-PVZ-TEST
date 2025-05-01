package audit

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockAuditLog реализует интерфейс AuditLog для тестирования
type MockAuditLog struct {
	OnLogPVZCreation func(ctx context.Context, pvzID, userID uuid.UUID) error
}

func (m *MockAuditLog) LogPVZCreation(ctx context.Context, pvzID, userID uuid.UUID) error {
	if m.OnLogPVZCreation != nil {
		return m.OnLogPVZCreation(ctx, pvzID, userID)
	}
	return nil
}

func TestAuditLog_Interface(t *testing.T) {
	// Проверяем, что MockAuditLog реализует интерфейс AuditLog
	var _ AuditLog = &MockAuditLog{}
}

func TestMockAuditLog_LogPVZCreation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*MockAuditLog, context.Context, uuid.UUID, uuid.UUID)
		wantErr bool
	}{
		{
			name: "успешное логирование создания ПВЗ",
			setup: func() (*MockAuditLog, context.Context, uuid.UUID, uuid.UUID) {
				auditLog := &MockAuditLog{
					OnLogPVZCreation: func(ctx context.Context, pvzID, userID uuid.UUID) error {
						return nil
					},
				}
				return auditLog, context.Background(), uuid.New(), uuid.New()
			},
			wantErr: false,
		},
		{
			name: "ошибка при логировании",
			setup: func() (*MockAuditLog, context.Context, uuid.UUID, uuid.UUID) {
				expectedErr := assert.AnError
				auditLog := &MockAuditLog{
					OnLogPVZCreation: func(ctx context.Context, pvzID, userID uuid.UUID) error {
						return expectedErr
					},
				}
				return auditLog, context.Background(), uuid.New(), uuid.New()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditLog, ctx, pvzID, userID := tt.setup()
			err := auditLog.LogPVZCreation(ctx, pvzID, userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
