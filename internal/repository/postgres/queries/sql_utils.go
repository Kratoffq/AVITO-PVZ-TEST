package queries

import (
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// PostgresBuilder возвращает SQL-билдер для PostgreSQL
var PostgresBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// FormatUUID форматирует UUID в строку
func FormatUUID(id uuid.UUID) string {
	return id.String()
}

// FormatUUIDs форматирует массив UUID в массив строк
func FormatUUIDs(ids []uuid.UUID) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}

// Paginate применяет пагинацию к SQL-запросу
func Paginate(builder squirrel.SelectBuilder, offset, limit int) (string, []interface{}, error) {
	return builder.
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()
}
