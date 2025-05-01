package queries

import (
	"github.com/Masterminds/squirrel"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/google/uuid"
)

// CreateReception создает новую приемку
func CreateReception(r *reception.Reception) (string, []interface{}, error) {
	return PostgresBuilder.Insert("receptions").
		Columns("id", "date_time", "pvz_id", "status").
		Values(FormatUUID(r.ID), r.DateTime, FormatUUID(r.PVZID), r.Status).
		ToSql()
}

// GetReceptionByID получает приемку по ID
func GetReceptionByID(id uuid.UUID) (string, []interface{}, error) {
	return PostgresBuilder.Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(squirrel.Eq{"id": FormatUUID(id)}).
		ToSql()
}

// GetOpenReceptionByPVZID получает открытую приемку для ПВЗ
func GetOpenReceptionByPVZID(pvzID uuid.UUID) (string, []interface{}, error) {
	return PostgresBuilder.Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(squirrel.Eq{
			"pvz_id": FormatUUID(pvzID),
			"status": reception.StatusInProgress,
		}).
		ToSql()
}

// CloseReception закрывает приемку
func CloseReception(id uuid.UUID) (string, []interface{}, error) {
	return PostgresBuilder.Update("receptions").
		Set("status", reception.StatusClose).
		Where(squirrel.Eq{"id": FormatUUID(id)}).
		ToSql()
}

// ListReceptions получает список приемок с пагинацией
func ListReceptions(offset, limit int) (string, []interface{}, error) {
	return Paginate(
		PostgresBuilder.Select("id", "date_time", "pvz_id", "status").
			From("receptions").
			OrderBy("date_time DESC"),
		offset,
		limit,
	)
}
