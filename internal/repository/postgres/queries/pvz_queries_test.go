package queries

import (
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/avito/pvz/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePVZQuery(t *testing.T) {
	id := uuid.New()
	dateTime := time.Now()
	city := "Moscow"

	query, args, err := CreatePVZ(id, dateTime, city)
	require.NoError(t, err)
	assert.Equal(t, "INSERT INTO pvzs (id,created_at,city) VALUES ($1,$2,$3)", query)
	assert.Len(t, args, 3)
	assert.Equal(t, id.String(), args[0])
	assert.Equal(t, dateTime, args[1])
	assert.Equal(t, city, args[2])
}

func TestGetPVZByIDQuery(t *testing.T) {
	id := uuid.New()
	query, args, err := GetPVZByID(id)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, created_at, city FROM pvzs WHERE id = $1", query)
	assert.Len(t, args, 1)
	assert.Equal(t, id.String(), args[0])
}

func TestUpdatePVZQuery(t *testing.T) {
	id := uuid.New()
	city := "Moscow"

	query, args, err := UpdatePVZ(id, city)
	require.NoError(t, err)
	assert.Equal(t, "UPDATE pvzs SET city = $1 WHERE id = $2", query)
	assert.Len(t, args, 2)
	assert.Equal(t, city, args[0])
	assert.Equal(t, id.String(), args[1])
}

func TestDeletePVZQuery(t *testing.T) {
	id := uuid.New()
	query, args, err := DeletePVZ(id)
	require.NoError(t, err)
	assert.Equal(t, "DELETE FROM pvzs WHERE id = $1", query)
	assert.Len(t, args, 1)
	assert.Equal(t, id.String(), args[0])
}

func TestListPVZsQuery(t *testing.T) {
	query, args, err := ListPVZs(20, 10)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, created_at, city FROM pvzs ORDER BY created_at DESC LIMIT 10 OFFSET 20", query)
	assert.Equal(t, []interface{}{}, args)
}

func TestGetPVZByCityQuery(t *testing.T) {
	city := "Moscow"
	query, args, err := GetPVZByCity(city)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, created_at, city FROM pvzs WHERE city = $1", query)
	assert.Len(t, args, 1)
	assert.Equal(t, city, args[0])
}

func TestCreatePVZ(t *testing.T) {
	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             "Москва",
	}

	query, args, err := PostgresBuilder.
		Insert("pvzs").
		Columns("id", "created_at", "city").
		Values(pvz.ID, pvz.RegistrationDate, pvz.City).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "INSERT INTO pvzs (id,created_at,city) VALUES ($1,$2,$3)", query)
	assert.Len(t, args, 3)
	assert.Equal(t, pvz.ID, args[0])
	assert.Equal(t, pvz.RegistrationDate, args[1])
	assert.Equal(t, pvz.City, args[2])
}

func TestGetPVZByID(t *testing.T) {
	id := uuid.New()
	query, args, err := PostgresBuilder.
		Select("id", "created_at", "city").
		From("pvzs").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, created_at, city FROM pvzs WHERE id = $1", query)
	assert.Equal(t, []interface{}{id.String()}, args)
}

func TestUpdatePVZ(t *testing.T) {
	id := uuid.New()
	query, args, err := PostgresBuilder.
		Update("pvzs").
		Set("city", "Санкт-Петербург").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "UPDATE pvzs SET city = $1 WHERE id = $2", query)
	assert.Equal(t, []interface{}{"Санкт-Петербург", id.String()}, args)
}

func TestDeletePVZ(t *testing.T) {
	id := uuid.New()
	query, args, err := PostgresBuilder.
		Delete("pvzs").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "DELETE FROM pvzs WHERE id = $1", query)
	assert.Equal(t, []interface{}{id.String()}, args)
}

func TestListPVZs(t *testing.T) {
	query, args, err := PostgresBuilder.
		Select("id", "created_at", "city").
		From("pvzs").
		OrderBy("created_at DESC").
		Limit(20).
		Offset(10).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, created_at, city FROM pvzs ORDER BY created_at DESC LIMIT 20 OFFSET 10", query)
	assert.Empty(t, args)
}

func TestGetPVZByCity(t *testing.T) {
	query, args, err := PostgresBuilder.
		Select("id", "created_at", "city").
		From("pvzs").
		Where(squirrel.Eq{"city": "Москва"}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, created_at, city FROM pvzs WHERE city = $1", query)
	assert.Equal(t, []interface{}{"Москва"}, args)
}
