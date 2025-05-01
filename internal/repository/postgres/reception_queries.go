package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/avito/pvz/internal/models"
	"github.com/google/uuid"
)

// ReceptionQueries содержит запросы для работы с приемками
type ReceptionQueries struct {
	db *sql.DB
}

// NewReceptionQueries создает новый экземпляр ReceptionQueries
func NewReceptionQueries(db *sql.DB) *ReceptionQueries {
	return &ReceptionQueries{db: db}
}

// Create создает новую приемку
func (q *ReceptionQueries) Create(ctx context.Context, reception *models.Reception) error {
	query := `INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, $2, $3, $4)`
	_, err := q.db.ExecContext(ctx, query, reception.ID, reception.DateTime, reception.PVZID, reception.Status)
	if err != nil {
		return fmt.Errorf("failed to create reception: %w", err)
	}
	return nil
}

// GetByID получает приемку по ID
func (q *ReceptionQueries) GetByID(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	query := `SELECT id, date_time, pvz_id, status FROM receptions WHERE id = $1`
	row := q.db.QueryRowContext(ctx, query, id)
	reception := &models.Reception{}
	err := row.Scan(&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get reception: %w", err)
	}
	return reception, nil
}

// UpdateStatus обновляет статус приемки
func (q *ReceptionQueries) UpdateStatus(ctx context.Context, id uuid.UUID, status models.ReceptionStatus) error {
	query := `UPDATE receptions SET status = $1 WHERE id = $2`
	result, err := q.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update reception status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("reception not found")
	}

	return nil
}

// GetProducts получает список товаров приемки
func (q *ReceptionQueries) GetProducts(ctx context.Context, receptionID uuid.UUID) ([]*models.Product, error) {
	query := `SELECT id, date_time, type, reception_id FROM products WHERE reception_id = $1 ORDER BY date_time DESC`
	rows, err := q.db.QueryContext(ctx, query, receptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		err := rows.Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	return products, nil
}
