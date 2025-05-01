package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/avito/pvz/internal/repository/postgres/queries"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ReceptionRepository реализует интерфейс reception.Repository
type ReceptionRepository struct {
	db *sqlx.DB
}

// NewReceptionRepository создает новый экземпляр ReceptionRepository
func NewReceptionRepository(db *sqlx.DB) *ReceptionRepository {
	return &ReceptionRepository{db: db}
}

// Create создает новую приемку
func (r *ReceptionRepository) Create(ctx context.Context, reception *reception.Reception) error {
	query, args, err := queries.CreateReception(reception)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// GetByID получает приемку по ID
func (r *ReceptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*reception.Reception, error) {
	query, args, err := queries.GetReceptionByID(id)
	if err != nil {
		return nil, err
	}

	var result reception.Reception
	err = r.db.GetContext(ctx, &result, query, args...)
	if err == sql.ErrNoRows {
		return nil, reception.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetOpenByPVZID получает открытую приемку для ПВЗ
func (r *ReceptionRepository) GetOpenByPVZID(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	query, args, err := queries.GetOpenReceptionByPVZID(pvzID)
	if err != nil {
		return nil, err
	}

	var result reception.Reception
	err = r.db.GetContext(ctx, &result, query, args...)
	if err == sql.ErrNoRows {
		return nil, reception.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Close закрывает приемку
func (r *ReceptionRepository) Close(ctx context.Context, id uuid.UUID) error {
	query, args, err := queries.CloseReception(id)
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
		return reception.ErrNotFound
	}

	return nil
}

// List возвращает список приемок с пагинацией
func (r *ReceptionRepository) List(ctx context.Context, offset, limit int) ([]*reception.Reception, error) {
	query, args, err := queries.ListReceptions(offset, limit)
	if err != nil {
		return nil, err
	}

	var result []*reception.Reception
	err = r.db.SelectContext(ctx, &result, query, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Update обновляет данные приемки
func (r *ReceptionRepository) Update(ctx context.Context, reception *reception.Reception) error {
	query := `UPDATE receptions SET status = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, reception.Status, reception.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("reception not found")
	}

	return nil
}

// Delete удаляет приёмку по ID
func (r *ReceptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM receptions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete reception: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return reception.ErrNotFound
	}

	return nil
}

// GetProducts получает все товары приемки
func (r *ReceptionRepository) GetProducts(ctx context.Context, receptionID uuid.UUID) ([]*product.Product, error) {
	query := `
		SELECT id, date_time, type, reception_id 
		FROM products 
		WHERE reception_id = $1 
		ORDER BY date_time DESC
	`
	var products []*product.Product
	err := r.db.SelectContext(ctx, &products, query, receptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	return products, nil
}

// GetLastOpen получает последнюю открытую приемку для ПВЗ
func (r *ReceptionRepository) GetLastOpen(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	query := `
		SELECT id, date_time, pvz_id, status 
		FROM receptions 
		WHERE pvz_id = $1 AND status = $2 
		ORDER BY date_time DESC 
		LIMIT 1
	`
	var result reception.Reception
	err := r.db.GetContext(ctx, &result, query, pvzID, reception.StatusInProgress)
	if err == sql.ErrNoRows {
		return nil, reception.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}
