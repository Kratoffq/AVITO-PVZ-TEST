package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/avito/pvz/internal/models"
	"github.com/avito/pvz/internal/repository"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// DB представляет соединение с базой данных
type DB struct {
	*sql.DB
}

// New создает новое соединение с базой данных
func New(cfg struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

type PostgresRepository struct {
	db    *sql.DB
	cache *Cache
}

func NewRepository(db *sql.DB) *PostgresRepository {
	// Настройка пула соединений
	db.SetMaxOpenConns(1000)               // Максимальное количество открытых соединений
	db.SetMaxIdleConns(100)                // Максимальное количество простаивающих соединений
	db.SetConnMaxLifetime(time.Hour)       // Максимальное время жизни соединения
	db.SetConnMaxIdleTime(time.Minute * 5) // Максимальное время простоя соединения

	return &PostgresRepository{
		db:    db,
		cache: NewCache(),
	}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id, email, password, role, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.Password, user.Role, user.CreatedAt)
	return err
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, email, password, role, created_at FROM users WHERE email = $1`
	row := r.db.QueryRowContext(ctx, query, email)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `SELECT id, email, password, role, created_at FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *PostgresRepository) CreatePVZ(ctx context.Context, pvz *models.PVZ) error {
	query := `INSERT INTO pvzs (id, registration_date, city) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, pvz.ID, pvz.RegistrationDate, pvz.City)
	if err != nil {
		return fmt.Errorf("failed to create PVZ: %w", err)
	}

	// Инвалидируем кэш
	r.cache.DeletePVZ(pvz.ID)
	return nil
}

func (r *PostgresRepository) GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	// Проверяем кэш
	if pvz, ok := r.cache.GetPVZ(id); ok {
		return pvz, nil
	}

	query := `SELECT id, registration_date, city FROM pvzs WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	pvz := &models.PVZ{}
	err := row.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get PVZ: %w", err)
	}

	// Сохраняем в кэш
	r.cache.SetPVZ(pvz)
	return pvz, nil
}

func (r *PostgresRepository) GetPVZsWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*repository.PVZWithReceptions, error) {
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
	rows, err := r.db.QueryContext(ctx, query, startDate, endDate, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
			return nil, err
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

func (r *PostgresRepository) CreateReception(ctx context.Context, reception *models.Reception) error {
	query := `INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, reception.ID, reception.DateTime, reception.PVZID, reception.Status)
	return err
}

func (r *PostgresRepository) GetReceptionByID(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	// Проверяем кэш
	if reception, ok := r.cache.GetReception(id); ok {
		return reception, nil
	}

	query := `SELECT id, date_time, pvz_id, status FROM receptions WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	reception := &models.Reception{}
	err := row.Scan(&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш
	r.cache.SetReception(reception)
	return reception, nil
}

func (r *PostgresRepository) GetLastOpenReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	query := `
		SELECT id, date_time, pvz_id, status 
		FROM receptions 
		WHERE pvz_id = $1 AND status = $2 
		ORDER BY date_time DESC 
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, pvzID, models.StatusInProgress)
	reception := &models.Reception{}
	err := row.Scan(&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return reception, err
}

func (r *PostgresRepository) CloseReception(ctx context.Context, receptionID uuid.UUID) error {
	query := `UPDATE receptions SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, models.StatusClose, receptionID)
	return err
}

func (r *PostgresRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	query := `INSERT INTO products (id, date_time, type, reception_id) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, product.ID, product.DateTime, product.Type, product.ReceptionID)
	return err
}

func (r *PostgresRepository) GetLastProduct(ctx context.Context, receptionID uuid.UUID) (*models.Product, error) {
	query := `
		SELECT id, date_time, type, reception_id 
		FROM products 
		WHERE reception_id = $1 
		ORDER BY date_time DESC 
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, receptionID)
	product := &models.Product{}
	err := row.Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш
	r.cache.SetProduct(product)
	return product, nil
}

func (r *PostgresRepository) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, productID)
	return err
}

func (r *PostgresRepository) CreateProducts(ctx context.Context, products []*models.Product) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO products (id, date_time, type, reception_id) 
		VALUES ($1, $2, $3, $4)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, product := range products {
		_, err := stmt.ExecContext(ctx,
			product.ID,
			product.DateTime,
			product.Type,
			product.ReceptionID,
		)
		if err != nil {
			return err
		}
		r.cache.SetProduct(product)
	}

	return tx.Commit()
}

func (r *PostgresRepository) UpdatePVZ(ctx context.Context, pvz *models.PVZ) error {
	query := `UPDATE pvzs SET city = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, pvz.City, pvz.ID)
	if err != nil {
		return fmt.Errorf("failed to update PVZ: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("PVZ not found")
	}

	// Инвалидируем кэш
	r.cache.DeletePVZ(pvz.ID)
	return nil
}

func (r *PostgresRepository) DeletePVZ(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM pvzs WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete PVZ: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("PVZ not found")
	}

	// Инвалидируем кэш
	r.cache.DeletePVZ(id)
	return nil
}
