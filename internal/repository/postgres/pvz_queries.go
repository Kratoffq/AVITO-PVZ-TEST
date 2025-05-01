package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/avito/pvz/internal/models"
	"github.com/avito/pvz/internal/repository"
	"github.com/google/uuid"
)

// PVZQueries содержит запросы для работы с ПВЗ
type PVZQueries struct {
	db *sql.DB
}

// NewPVZQueries создает новый экземпляр PVZQueries
func NewPVZQueries(db *sql.DB) *PVZQueries {
	return &PVZQueries{db: db}
}

// Create создает новый ПВЗ
func (q *PVZQueries) Create(ctx context.Context, pvz *models.PVZ) error {
	// Валидация города
	if pvz.City != "Москва" {
		return fmt.Errorf("invalid city: %s, only 'Москва' is allowed", pvz.City)
	}

	query := `INSERT INTO pvzs (id, registration_date, city) VALUES ($1, $2, $3)`
	_, err := q.db.ExecContext(ctx, query, pvz.ID, pvz.RegistrationDate, pvz.City)
	if err != nil {
		return fmt.Errorf("failed to create PVZ: %w", err)
	}
	return nil
}

// GetByID получает ПВЗ по ID
func (q *PVZQueries) GetByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	query := `SELECT id, registration_date, city FROM pvzs WHERE id = $1`
	row := q.db.QueryRowContext(ctx, query, id)
	pvz := &models.PVZ{}
	err := row.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get PVZ: %w", err)
	}
	return pvz, nil
}

// GetWithReceptions получает список ПВЗ с приемками
func (q *PVZQueries) GetWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*repository.PVZWithReceptions, error) {
	offset := (page - 1) * limit
	query := `
		WITH filtered_receptions AS (
			SELECT r.*, p.id as product_id, p.date_time as product_date_time, p.type as product_type
			FROM receptions r
			LEFT JOIN products p ON r.id = p.reception_id
			WHERE r.date_time BETWEEN $1 AND $2
		)
		SELECT 
			p.id as pvz_id,
			p.registration_date,
			p.city,
			r.id as reception_id,
			r.date_time as reception_date_time,
			r.pvz_id,
			r.status,
			r.product_id,
			r.product_date_time,
			r.product_type
		FROM pvzs p
		LEFT JOIN filtered_receptions r ON p.id = r.pvz_id
		ORDER BY p.registration_date DESC
		LIMIT $3 OFFSET $4`

	rows, err := q.db.QueryContext(ctx, query, startDate, endDate, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get PVZs with receptions: %w", err)
	}
	defer rows.Close()

	return q.scanPVZsWithReceptions(rows)
}

// scanPVZsWithReceptions сканирует результаты запроса в структуру PVZWithReceptions
func (q *PVZQueries) scanPVZsWithReceptions(rows *sql.Rows) ([]*repository.PVZWithReceptions, error) {
	pvzs := make(map[uuid.UUID]*repository.PVZWithReceptions)
	receptions := make(map[uuid.UUID]*repository.ReceptionWithProducts)

	for rows.Next() {
		var pvzID, receptionID, productID uuid.UUID
		var registrationDate, receptionDateTime, productDateTime time.Time
		var city string
		var receptionStatus models.ReceptionStatus
		var productType models.ProductType

		err := rows.Scan(
			&pvzID, &registrationDate, &city,
			&receptionID, &receptionDateTime, &pvzID, &receptionStatus,
			&productID, &productDateTime, &productType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan PVZ with receptions: %w", err)
		}

		if _, ok := pvzs[pvzID]; !ok {
			pvzs[pvzID] = &repository.PVZWithReceptions{
				PVZ: &models.PVZ{
					ID:               pvzID,
					RegistrationDate: registrationDate,
					City:             city,
				},
				Receptions: make([]*repository.ReceptionWithProducts, 0),
			}
		}

		if _, ok := receptions[receptionID]; !ok {
			receptions[receptionID] = &repository.ReceptionWithProducts{
				Reception: &models.Reception{
					ID:       receptionID,
					DateTime: receptionDateTime,
					PVZID:    pvzID,
					Status:   receptionStatus,
				},
				Products: make([]*models.Product, 0),
			}
			pvzs[pvzID].Receptions = append(pvzs[pvzID].Receptions, receptions[receptionID])
		}

		if productID != uuid.Nil {
			receptions[receptionID].Products = append(receptions[receptionID].Products, &models.Product{
				ID:          productID,
				DateTime:    productDateTime,
				Type:        productType,
				ReceptionID: productID,
			})
		}
	}

	result := make([]*repository.PVZWithReceptions, 0, len(pvzs))
	for _, pvz := range pvzs {
		result = append(result, pvz)
	}

	return result, nil
}

// GetAllPVZs получает список всех ПВЗ
func (q *PVZQueries) GetAllPVZs(ctx context.Context) ([]*models.PVZ, error) {
	query := `
		SELECT id, registration_date, city
		FROM pvzs
		ORDER BY registration_date DESC
	`

	rows, err := q.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pvzs []*models.PVZ
	for rows.Next() {
		var pvz models.PVZ
		if err := rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City); err != nil {
			return nil, err
		}
		pvzs = append(pvzs, &pvz)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pvzs, nil
}
