package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/avito/pvz/internal/repository/postgres/queries"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ProductRepository реализует интерфейс product.Repository
type ProductRepository struct {
	db *sqlx.DB
}

// NewProductRepository создает новый экземпляр ProductRepository
func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create создает новый товар
func (r *ProductRepository) Create(ctx context.Context, p *product.Product) error {
	query, args, err := queries.CreateProduct(p.ID, p.DateTime, string(p.Type), p.ReceptionID)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// CreateBatch создает несколько товаров в одной транзакции
func (r *ProductRepository) CreateBatch(ctx context.Context, products []*product.Product) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
			}
		}
	}()

	for _, p := range products {
		query, args, err := queries.CreateProduct(p.ID, p.DateTime, string(p.Type), p.ReceptionID)
		if err != nil {
			return fmt.Errorf("failed to create product query: %w", err)
		}

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to execute product creation: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID получает товар по ID
func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	query, args, err := queries.GetProductByID(id)
	if err != nil {
		return nil, err
	}

	var result product.Product
	err = r.db.GetContext(ctx, &result, query, args...)
	if err == sql.ErrNoRows {
		return nil, product.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetByReceptionID получает все товары приемки
func (r *ProductRepository) GetByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]*product.Product, error) {
	query, args, err := queries.GetProductsByReceptionID(receptionID)
	if err != nil {
		return nil, err
	}

	var result []*product.Product
	err = r.db.SelectContext(ctx, &result, query, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteLast удаляет последний добавленный товар приемки
func (r *ProductRepository) DeleteLast(ctx context.Context, receptionID uuid.UUID) error {
	query, args, err := queries.DeleteLastProduct(receptionID)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return product.ErrNotFound
	}

	return nil
}

// List возвращает список товаров с пагинацией
func (r *ProductRepository) List(ctx context.Context, offset, limit int) ([]*product.Product, error) {
	query, args, err := queries.ListProducts(offset, limit)
	if err != nil {
		return nil, err
	}

	var result []*product.Product
	err = r.db.SelectContext(ctx, &result, query, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Delete удаляет товар по ID
func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return product.ErrNotFound
	}

	return nil
}

// GetLast получает последний добавленный товар из приемки
func (r *ProductRepository) GetLast(ctx context.Context, receptionID uuid.UUID) (*product.Product, error) {
	query := `
		SELECT id, date_time, type, reception_id 
		FROM products 
		WHERE reception_id = $1 
		ORDER BY date_time DESC 
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, receptionID)
	product := &product.Product{}
	err := row.Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return product, err
}

// Update обновляет информацию о товаре
func (r *ProductRepository) Update(ctx context.Context, p *product.Product) error {
	query := `
		UPDATE products 
		SET type = $1, date_time = $2 
		WHERE id = $3
	`
	result, err := r.db.ExecContext(ctx, query, p.Type, p.DateTime, p.ID)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return product.ErrNotFound
	}

	return nil
}
