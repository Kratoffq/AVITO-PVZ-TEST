package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/avito/pvz/internal/models"
	"github.com/google/uuid"
)

// ProductQueries содержит запросы для работы с товарами
type ProductQueries struct {
	db *sql.DB
}

// NewProductQueries создает новый экземпляр ProductQueries
func NewProductQueries(db *sql.DB) *ProductQueries {
	return &ProductQueries{db: db}
}

// Create создает новый товар
func (q *ProductQueries) Create(ctx context.Context, product *models.Product) error {
	query := `INSERT INTO products (id, date_time, type, reception_id) VALUES ($1, $2, $3, $4)`
	_, err := q.db.ExecContext(ctx, query, product.ID, product.DateTime, product.Type, product.ReceptionID)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

// CreateBatch создает несколько товаров в одной транзакции
func (q *ProductQueries) CreateBatch(ctx context.Context, products []*models.Product) error {
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO products (id, date_time, type, reception_id) VALUES ($1, $2, $3, $4)`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, product := range products {
		_, err := stmt.ExecContext(ctx, product.ID, product.DateTime, product.Type, product.ReceptionID)
		if err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteLast удаляет последний добавленный товар из приемки
func (q *ProductQueries) DeleteLast(ctx context.Context, receptionID uuid.UUID) error {
	query := `
		DELETE FROM products 
		WHERE id = (
			SELECT id 
			FROM products 
			WHERE reception_id = $1 
			ORDER BY date_time DESC 
			LIMIT 1
		)`
	result, err := q.db.ExecContext(ctx, query, receptionID)
	if err != nil {
		return fmt.Errorf("failed to delete last product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no products found in reception")
	}

	return nil
}

// GetByReceptionID получает все товары приемки
func (q *ProductQueries) GetByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]*models.Product, error) {
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
