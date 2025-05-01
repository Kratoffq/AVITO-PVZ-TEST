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
	registrationDate := time.Now()
	city := "Москва"

	query, args, err := CreatePVZ(id, registrationDate, city)
	require.NoError(t, err)
	assert.Equal(t, "INSERT INTO pvzs (id,registration_date,city) VALUES ($1,$2,$3)", query)
	assert.Len(t, args, 3)
	assert.Equal(t, id, args[0])
	assert.Equal(t, registrationDate, args[1])
	assert.Equal(t, city, args[2])
}

func TestGetPVZByIDQuery(t *testing.T) {
	id := uuid.New()
	query, args, err := GetPVZByID(id)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, registration_date, city FROM pvzs WHERE id = $1", query)
	assert.Equal(t, []interface{}{id}, args)
}

func TestUpdatePVZQuery(t *testing.T) {
	id := uuid.New()
	city := "Санкт-Петербург"
	query, args, err := UpdatePVZ(id, city)
	require.NoError(t, err)
	assert.Equal(t, "UPDATE pvzs SET city = $1 WHERE id = $2", query)
	assert.Equal(t, []interface{}{city, id}, args)
}

func TestDeletePVZQuery(t *testing.T) {
	id := uuid.New()
	query, args, err := DeletePVZ(id)
	require.NoError(t, err)
	assert.Equal(t, "DELETE FROM pvzs WHERE id = $1", query)
	assert.Equal(t, []interface{}{id}, args)
}

func TestListPVZsQuery(t *testing.T) {
	offset, limit := 10, 20
	query, args, err := ListPVZs(offset, limit)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, registration_date, city FROM pvzs ORDER BY registration_date DESC LIMIT $2 OFFSET $1", query)
	assert.Equal(t, []interface{}{offset, limit}, args)
}

func TestGetPVZByCityQuery(t *testing.T) {
	city := "Москва"
	query, args, err := GetPVZByCity(city)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, registration_date, city FROM pvzs WHERE city = $1", query)
	assert.Equal(t, []interface{}{city}, args)
}

func TestCreatePVZ(t *testing.T) {
	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             "Москва",
	}

	query, args, err := PostgresBuilder.
		Insert("pvzs").
		Columns("id", "registration_date", "city").
		Values(pvz.ID, pvz.RegistrationDate, pvz.City).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "INSERT INTO pvzs (id,registration_date,city) VALUES ($1,$2,$3)", query)
	assert.Len(t, args, 3)
	assert.Equal(t, pvz.ID, args[0])
	assert.Equal(t, pvz.RegistrationDate, args[1])
	assert.Equal(t, pvz.City, args[2])
}

func TestGetPVZByID(t *testing.T) {
	id := uuid.New()
	query, args, err := PostgresBuilder.
		Select("id", "registration_date", "city").
		From("pvzs").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, registration_date, city FROM pvzs WHERE id = $1", query)
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
		Select("id", "registration_date", "city").
		From("pvzs").
		OrderBy("registration_date DESC").
		Limit(20).
		Offset(10).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, registration_date, city FROM pvzs ORDER BY registration_date DESC LIMIT 20 OFFSET 10", query)
	assert.Empty(t, args)
}

func TestGetPVZByCity(t *testing.T) {
	query, args, err := PostgresBuilder.
		Select("id", "registration_date", "city").
		From("pvzs").
		Where(squirrel.Eq{"city": "Москва"}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, registration_date, city FROM pvzs WHERE city = $1", query)
	assert.Equal(t, []interface{}{"Москва"}, args)
}
