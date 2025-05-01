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

func TestCreateProductQuery(t *testing.T) {
	id := uuid.New()
	dateTime := time.Now()
	productType := string(models.TypeElectronics)
	receptionID := uuid.New()

	query, args, err := CreateProduct(id, dateTime, productType, receptionID)
	require.NoError(t, err)
	assert.Equal(t, "INSERT INTO products (id,date_time,type,reception_id) VALUES ($1,$2,$3,$4)", query)
	assert.Len(t, args, 4)
	assert.Equal(t, id, args[0])
	assert.Equal(t, dateTime, args[1])
	assert.Equal(t, productType, args[2])
	assert.Equal(t, receptionID, args[3])
}

func TestGetProductByIDQuery(t *testing.T) {
	id := uuid.New()
	query, args, err := GetProductByID(id)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, type, reception_id FROM products WHERE id = $1", query)
	assert.Equal(t, []interface{}{id}, args)
}

func TestGetProductsByReceptionIDQuery(t *testing.T) {
	receptionID := uuid.New()
	query, args, err := GetProductsByReceptionID(receptionID)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, type, reception_id FROM products WHERE reception_id = $1 ORDER BY date_time DESC", query)
	assert.Equal(t, []interface{}{receptionID}, args)
}

func TestDeleteLastProductQuery(t *testing.T) {
	receptionID := uuid.New()
	query, args, err := DeleteLastProduct(receptionID)
	require.NoError(t, err)
	assert.Equal(t, "DELETE FROM products WHERE id IN (SELECT id FROM products WHERE reception_id = $1 ORDER BY date_time DESC LIMIT 1)", query)
	assert.Equal(t, []interface{}{receptionID}, args)
}

func TestListProductsQuery(t *testing.T) {
	offset, limit := 10, 20
	query, args, err := ListProducts(offset, limit)
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, type, reception_id FROM products ORDER BY date_time DESC LIMIT $2 OFFSET $1", query)
	assert.Equal(t, []interface{}{offset, limit}, args)
}

func TestCreateProduct(t *testing.T) {
	product := &models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        models.TypeElectronics,
		ReceptionID: uuid.New(),
	}

	query, args, err := PostgresBuilder.
		Insert("products").
		Columns("id", "date_time", "type", "reception_id").
		Values(product.ID, product.DateTime, product.Type, product.ReceptionID).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "INSERT INTO products (id,date_time,type,reception_id) VALUES ($1,$2,$3,$4)", query)
	assert.Len(t, args, 4)
	assert.Equal(t, product.ID, args[0])
	assert.Equal(t, product.DateTime, args[1])
	assert.Equal(t, product.Type, args[2])
	assert.Equal(t, product.ReceptionID, args[3])
}

func TestListProducts(t *testing.T) {
	receptionID := uuid.New()
	query, args, err := PostgresBuilder.
		Select("id", "date_time", "type", "reception_id").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionID}).
		OrderBy("date_time DESC").
		Limit(20).
		Offset(10).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, type, reception_id FROM products WHERE reception_id = $1 ORDER BY date_time DESC LIMIT 20 OFFSET 10", query)
	assert.Equal(t, []interface{}{receptionID.String()}, args)
}

func TestDeleteLastProduct(t *testing.T) {
	receptionID := uuid.New()
	query, args, err := PostgresBuilder.
		Delete("products").
		Where(squirrel.Eq{"reception_id": receptionID}).
		OrderBy("date_time DESC").
		Limit(1).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "DELETE FROM products WHERE reception_id = $1 ORDER BY date_time DESC LIMIT 1", query)
	assert.Equal(t, []interface{}{receptionID.String()}, args)
}

func TestGetProductByID(t *testing.T) {
	productID := uuid.New()
	query, args, err := PostgresBuilder.
		Select("id", "date_time", "type", "reception_id").
		From("products").
		Where(squirrel.Eq{"id": productID}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, type, reception_id FROM products WHERE id = $1", query)
	assert.Equal(t, []interface{}{productID.String()}, args)
}

func TestGetProductsByType(t *testing.T) {
	receptionID := uuid.New()
	productType := models.TypeElectronics
	query, args, err := PostgresBuilder.
		Select("id", "date_time", "type", "reception_id").
		From("products").
		Where(squirrel.Eq{
			"reception_id": receptionID,
			"type":         productType,
		}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, date_time, type, reception_id FROM products WHERE reception_id = $1 AND type = $2", query)
	assert.Equal(t, []interface{}{receptionID.String(), productType}, args)
}
