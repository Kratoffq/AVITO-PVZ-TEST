package queries

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// CreateProduct создает новый товар
func CreateProduct(id uuid.UUID, dateTime time.Time, productType string, receptionID uuid.UUID) (string, []interface{}, error) {
	return PostgresBuilder.Insert("products").
		Columns("id", "date_time", "type", "reception_id").
		Values(FormatUUID(id), dateTime, productType, FormatUUID(receptionID)).
		ToSql()
}

// GetProductByID получает товар по ID
func GetProductByID(id uuid.UUID) (string, []interface{}, error) {
	return PostgresBuilder.Select("id", "date_time", "type", "reception_id").
		From("products").
		Where(squirrel.Eq{"id": FormatUUID(id)}).
		ToSql()
}

// GetProductsByReceptionID получает товары по ID приемки
func GetProductsByReceptionID(receptionID uuid.UUID) (string, []interface{}, error) {
	return PostgresBuilder.Select("id", "date_time", "type", "reception_id").
		From("products").
		Where(squirrel.Eq{"reception_id": FormatUUID(receptionID)}).
		OrderBy("date_time DESC").
		ToSql()
}

// DeleteLastProduct удаляет последний товар по ID приемки
func DeleteLastProduct(receptionID uuid.UUID) (string, []interface{}, error) {
	return PostgresBuilder.Delete("products").
		Where("id = (SELECT id FROM products WHERE reception_id = ? ORDER BY date_time DESC LIMIT 1)", FormatUUID(receptionID)).
		ToSql()
}

// ListProducts получает список товаров с пагинацией
func ListProducts(offset, limit int) (string, []interface{}, error) {
	return Paginate(
		PostgresBuilder.Select("id", "date_time", "type", "reception_id").
			From("products").
			OrderBy("date_time DESC"),
		offset,
		limit,
	)
}
