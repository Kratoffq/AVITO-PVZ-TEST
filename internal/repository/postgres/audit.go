package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// AuditLog реализует интерфейс audit.AuditLog для PostgreSQL
type AuditLog struct {
	db *sqlx.DB
}

// NewAuditLog создает новый экземпляр AuditLog
func NewAuditLog(db *sqlx.DB) *AuditLog {
	return &AuditLog{db: db}
}

// LogPVZCreation логирует создание ПВЗ
func (a *AuditLog) LogPVZCreation(ctx context.Context, pvzID, userID uuid.UUID) error {
	query := `
		INSERT INTO audit_logs (operation_type, entity_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := a.db.ExecContext(ctx, query, "pvz_creation", pvzID, userID, time.Now())
	return err
}
