package queries

import (
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateReceptionQuery(t *testing.T) {
	reception := &reception.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    uuid.New(),
		Status:   reception.StatusInProgress,
	}

	query, args, err := CreateReception(reception)
	require.NoError(t, err)
	assert.Equal(t, "INSERT INTO receptions (id,date_time,pvz_id,status) VALUES ($1,$2,$3,$4)", query)
	assert.Len(t, args, 4)
	assert.Equal(t, reception.ID, args[0])
	assert.Equal(t, reception.DateTime, args[1])
	assert.Equal(t, reception.PVZID, args[2])
	assert.Equal(t, reception.Status, args[3])
}

func TestGetReceptionByIDQuery(t *testing.T) {
	id := uuid.New()
	query, args, err := GetReceptionByID(id)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, pvz_id, status FROM receptions WHERE id = $1", query)
	assert.Equal(t, []interface{}{id}, args)
}

func TestGetOpenReceptionByPVZIDQuery(t *testing.T) {
	pvzID := uuid.New()
	query, args, err := GetOpenReceptionByPVZID(pvzID)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, pvz_id, status FROM receptions WHERE pvz_id = $1 AND status = $2", query)
	assert.Equal(t, []interface{}{pvzID, reception.StatusInProgress}, args)
}

func TestCloseReceptionQuery(t *testing.T) {
	id := uuid.New()
	query, args, err := CloseReception(id)
	require.NoError(t, err)
	assert.Equal(t, "UPDATE receptions SET status = $1 WHERE id = $2", query)
	assert.Equal(t, []interface{}{reception.StatusClose, id}, args)
}

func TestListReceptionsQuery(t *testing.T) {
	offset, limit := 10, 20
	query, args, err := ListReceptions(offset, limit)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, pvz_id, status FROM receptions ORDER BY date_time DESC LIMIT $2 OFFSET $1", query)
	assert.Equal(t, []interface{}{offset, limit}, args)
}

func TestCreateReception(t *testing.T) {
	reception := &reception.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    uuid.New(),
		Status:   reception.StatusInProgress,
	}

	query, args, err := PostgresBuilder.
		Insert("receptions").
		Columns("id", "date_time", "pvz_id", "status").
		Values(reception.ID, reception.DateTime, reception.PVZID, reception.Status).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "INSERT INTO receptions (id,date_time,pvz_id,status) VALUES ($1,$2,$3,$4)", query)
	assert.Len(t, args, 4)
	assert.Equal(t, reception.ID, args[0])
	assert.Equal(t, reception.DateTime, args[1])
	assert.Equal(t, reception.PVZID, args[2])
	assert.Equal(t, reception.Status, args[3])
}

func TestGetReceptionByID(t *testing.T) {
	id := uuid.New()
	query, args, err := PostgresBuilder.
		Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, pvz_id, status FROM receptions WHERE id = $1", query)
	assert.Equal(t, []interface{}{id.String()}, args)
}

func TestGetOpenReceptionByPVZID(t *testing.T) {
	pvzID := uuid.New()
	query, args, err := PostgresBuilder.
		Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(squirrel.Eq{
			"pvz_id": pvzID,
			"status": reception.StatusInProgress,
		}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, pvz_id, status FROM receptions WHERE pvz_id = $1 AND status = $2", query)
	assert.Equal(t, []interface{}{pvzID.String(), reception.StatusInProgress}, args)
}

func TestCloseReception(t *testing.T) {
	id := uuid.New()
	query, args, err := PostgresBuilder.
		Update("receptions").
		Set("status", reception.StatusClose).
		Where(squirrel.Eq{"id": id}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "UPDATE receptions SET status = $1 WHERE id = $2", query)
	assert.Equal(t, []interface{}{reception.StatusClose, id.String()}, args)
}

func TestListReceptions(t *testing.T) {
	query, args, err := PostgresBuilder.
		Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		OrderBy("date_time DESC").
		Limit(20).
		Offset(10).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, pvz_id, status FROM receptions ORDER BY date_time DESC LIMIT 20 OFFSET 10", query)
	assert.Empty(t, args)
}
