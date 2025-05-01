package queries

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// CreatePVZ создает новый ПВЗ
func CreatePVZ(id uuid.UUID, createdAt time.Time, city string) (string, []interface{}, error) {
	return PostgresBuilder.Insert("pvzs").
		Columns("id", "registration_date", "city").
		Values(FormatUUID(id), createdAt, city).
		ToSql()
}

// GetPVZByID получает ПВЗ по ID
func GetPVZByID(id uuid.UUID) (string, []interface{}, error) {
	return PostgresBuilder.Select("id", "registration_date", "city").
		From("pvzs").
		Where(squirrel.Eq{"id": FormatUUID(id)}).
		ToSql()
}

// UpdatePVZ обновляет город ПВЗ
func UpdatePVZ(id uuid.UUID, city string) (string, []interface{}, error) {
	return PostgresBuilder.Update("pvzs").
		Set("city", city).
		Where(squirrel.Eq{"id": FormatUUID(id)}).
		ToSql()
}

// DeletePVZ удаляет ПВЗ
func DeletePVZ(id uuid.UUID) (string, []interface{}, error) {
	return PostgresBuilder.Delete("pvzs").
		Where(squirrel.Eq{"id": FormatUUID(id)}).
		ToSql()
}

// ListPVZs получает список ПВЗ с пагинацией
func ListPVZs(offset, limit int) (string, []interface{}, error) {
	return Paginate(
		PostgresBuilder.Select("id", "registration_date", "city").
			From("pvzs").
			OrderBy("registration_date DESC"),
		offset,
		limit,
	)
}

// GetPVZByCity получает ПВЗ по городу
func GetPVZByCity(city string) (string, []interface{}, error) {
	return PostgresBuilder.Select("id", "registration_date", "city").
		From("pvzs").
		Where(squirrel.Eq{"city": city}).
		ToSql()
}
