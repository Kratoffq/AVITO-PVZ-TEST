package audit

import (
	"context"

	"github.com/google/uuid"
)

// AuditLog определяет интерфейс для аудита операций
type AuditLog interface {
	// LogPVZCreation логирует создание ПВЗ
	LogPVZCreation(ctx context.Context, pvzID, userID uuid.UUID) error
}
