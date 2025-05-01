package queries

import (
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPostgresBuilder(t *testing.T) {
	// Проверяем, что PostgresBuilder использует правильный формат плейсхолдеров
	query, args, err := PostgresBuilder.Select("id").From("users").Where(squirrel.Eq{"id": 1}).ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT id FROM users WHERE id = $1", query)
	assert.Equal(t, []interface{}{1}, args)

	// Проверяем множественные условия
	query, args, err = PostgresBuilder.Select("id").From("users").
		Where(squirrel.Eq{"id": 1}).
		Where(squirrel.Eq{"name": "test"}).
		ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT id FROM users WHERE id = $1 AND name = $2", query)
	assert.Equal(t, []interface{}{1, "test"}, args)
}

func TestFormatUUID(t *testing.T) {
	id := uuid.New()
	formatted := FormatUUID(id)
	assert.Equal(t, id.String(), formatted)
}

func TestFormatUUIDs(t *testing.T) {
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	formatted := FormatUUIDs(ids)
	assert.Equal(t, len(ids), len(formatted))
	for i, id := range ids {
		assert.Equal(t, id.String(), formatted[i])
	}
}
