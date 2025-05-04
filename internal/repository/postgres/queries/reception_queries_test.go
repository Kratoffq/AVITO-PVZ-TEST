package queries

import (
	"testing"
	"time"

	"github.com/avito/pvz/internal/domain/reception"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateReceptionQuery(t *testing.T) {
	r := &reception.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    uuid.New(),
		Status:   reception.StatusInProgress,
	}

	query, args, err := CreateReception(r)
	require.NoError(t, err)
	assert.Equal(t, "INSERT INTO receptions (id,date_time,pvz_id,status) VALUES ($1,$2,$3,$4)", query)
	assert.Len(t, args, 4)
	assert.Equal(t, r.ID.String(), args[0])
	assert.Equal(t, r.DateTime, args[1])
	assert.Equal(t, r.PVZID.String(), args[2])
	assert.Equal(t, r.Status, args[3])
}

func TestGetReceptionByIDQuery(t *testing.T) {
	id := uuid.New()
	query, args, err := GetReceptionByID(id)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, pvz_id, status FROM receptions WHERE id = $1", query)
	assert.Len(t, args, 1)
	assert.Equal(t, id.String(), args[0])
}

func TestGetOpenReceptionByPVZIDQuery(t *testing.T) {
	pvzID := uuid.New()
	query, args, err := GetOpenReceptionByPVZID(pvzID)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, pvz_id, status FROM receptions WHERE pvz_id = $1 AND status = $2", query)
	assert.Len(t, args, 2)
	assert.Equal(t, pvzID.String(), args[0])
	assert.Equal(t, reception.StatusInProgress, args[1])
}

func TestCloseReceptionQuery(t *testing.T) {
	id := uuid.New()
	query, args, err := CloseReception(id)
	require.NoError(t, err)
	assert.Equal(t, "UPDATE receptions SET status = $1 WHERE id = $2", query)
	assert.Len(t, args, 2)
	assert.Equal(t, reception.StatusClose, args[0])
	assert.Equal(t, id.String(), args[1])
}

func TestListReceptionsQuery(t *testing.T) {
	query, args, err := ListReceptions(20, 10)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, pvz_id, status FROM receptions ORDER BY date_time DESC LIMIT 10 OFFSET 20", query)
	assert.Equal(t, []interface{}{}, args)
}
