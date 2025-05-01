package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/avito/pvz/internal/domain/product"
	domainpvz "github.com/avito/pvz/internal/domain/pvz"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/avito/pvz/internal/repository/postgres/queries"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// PVZRepository реализует интерфейс pvz.Repository
type PVZRepository struct {
	db *sqlx.DB
}

// NewPVZRepository создает новый экземпляр PVZRepository
func NewPVZRepository(db *sqlx.DB) *PVZRepository {
	return &PVZRepository{db: db}
}

// Create создает новый ПВЗ
func (r *PVZRepository) Create(ctx context.Context, pvz *domainpvz.PVZ) error {
	query, args, err := queries.CreatePVZ(pvz.ID, pvz.CreatedAt, pvz.City)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// GetByID получает ПВЗ по ID
func (r *PVZRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainpvz.PVZ, error) {
	query, args, err := queries.GetPVZByID(id)
	if err != nil {
		return nil, err
	}

	var result domainpvz.PVZ
	err = r.db.GetContext(ctx, &result, query, args...)
	if err == sql.ErrNoRows {
		return nil, domainpvz.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Update обновляет данные ПВЗ
func (r *PVZRepository) Update(ctx context.Context, pvz *domainpvz.PVZ) error {
	query, args, err := queries.UpdatePVZ(pvz.ID, pvz.City)
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
		return domainpvz.ErrNotFound
	}

	return nil
}

// Delete удаляет ПВЗ
func (r *PVZRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := queries.DeletePVZ(id)
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
		return domainpvz.ErrNotFound
	}

	return nil
}

// List возвращает список ПВЗ с пагинацией
func (r *PVZRepository) List(ctx context.Context, offset, limit int) ([]*domainpvz.PVZ, error) {
	query, args, err := queries.ListPVZs(offset, limit)
	if err != nil {
		return nil, err
	}

	var result []*domainpvz.PVZ
	err = r.db.SelectContext(ctx, &result, query, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetByCity получает ПВЗ по городу
func (r *PVZRepository) GetByCity(ctx context.Context, city string) (*domainpvz.PVZ, error) {
	query, args, err := queries.GetPVZByCity(city)
	if err != nil {
		return nil, err
	}

	var result domainpvz.PVZ
	err = r.db.GetContext(ctx, &result, query, args...)
	if err == sql.ErrNoRows {
		return nil, domainpvz.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetWithReceptions получает список ПВЗ с приемками за период
func (r *PVZRepository) GetWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*domainpvz.PVZWithReceptions, error) {
	offset := (page - 1) * limit
	query := `
		SELECT p.id, p.created_at, p.city, r.id, r.date_time, r.status
		FROM pvzs p
		LEFT JOIN receptions r ON p.id = r.pvz_id
		WHERE r.date_time BETWEEN $1 AND $2
		ORDER BY p.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pvzs := make(map[uuid.UUID]*domainpvz.PVZWithReceptions)
	for rows.Next() {
		var p domainpvz.PVZ
		var receptionID uuid.UUID
		var receptionDateTime time.Time
		var receptionStatus string
		err := rows.Scan(&p.ID, &p.CreatedAt, &p.City, &receptionID, &receptionDateTime, &receptionStatus)
		if err != nil {
			return nil, err
		}

		if _, ok := pvzs[p.ID]; !ok {
			pvzs[p.ID] = &domainpvz.PVZWithReceptions{
				PVZ:        &p,
				Receptions: make([]*domainpvz.ReceptionWithProducts, 0),
			}
		}

		if receptionID != uuid.Nil {
			reception := &domainpvz.ReceptionWithProducts{
				Reception: &reception.Reception{
					ID:       receptionID,
					DateTime: receptionDateTime,
					Status:   reception.Status(receptionStatus),
				},
				Products: make([]*product.Product, 0),
			}
			pvzs[p.ID].Receptions = append(pvzs[p.ID].Receptions, reception)
		}
	}

	result := make([]*domainpvz.PVZWithReceptions, 0, len(pvzs))
	for _, p := range pvzs {
		result = append(result, p)
	}

	return result, nil
}

// GetAll возвращает список всех ПВЗ
func (r *PVZRepository) GetAll(ctx context.Context) ([]*domainpvz.PVZ, error) {
	query := `SELECT id, city, created_at FROM pvzs ORDER BY created_at DESC`

	var pvzs []*domainpvz.PVZ
	err := r.db.SelectContext(ctx, &pvzs, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all PVZs: %w", err)
	}

	return pvzs, nil
}
