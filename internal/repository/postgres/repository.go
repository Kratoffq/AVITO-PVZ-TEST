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

// Repository реализует интерфейс repository.Repository
type Repository struct {
	db               *sql.DB
	pvzQueries       *PVZQueries
	receptionQueries *ReceptionQueries
	productQueries   *ProductQueries
}

// NewMainRepository создает новый экземпляр Repository
func NewMainRepository(db *sql.DB) repository.Repository {
	return &Repository{
		db:               db,
		pvzQueries:       NewPVZQueries(db),
		receptionQueries: NewReceptionQueries(db),
		productQueries:   NewProductQueries(db),
	}
}

// WithTx выполняет функцию в транзакции с настройками изоляции
func (r *Repository) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	// Устанавливаем уровень изоляции SERIALIZABLE для максимальной консистентности
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CreatePVZ создает новый ПВЗ
func (r *Repository) CreatePVZ(ctx context.Context, pvz *models.PVZ) error {
	return r.pvzQueries.Create(ctx, pvz)
}

// GetPVZByID получает ПВЗ по ID
func (r *Repository) GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	return r.pvzQueries.GetByID(ctx, id)
}

// GetPVZsWithReceptions получает список ПВЗ с приемками
func (r *Repository) GetPVZsWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*repository.PVZWithReceptions, error) {
	return r.pvzQueries.GetWithReceptions(ctx, startDate, endDate, page, limit)
}

// CreateReception создает новую приемку
func (r *Repository) CreateReception(ctx context.Context, reception *models.Reception) error {
	return r.receptionQueries.Create(ctx, reception)
}

// GetReceptionByID получает приемку по ID с блокировкой
func (r *Repository) GetReceptionByID(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	return r.receptionQueries.GetByID(ctx, id)
}

// UpdateReceptionStatus обновляет статус приемки
func (r *Repository) UpdateReceptionStatus(ctx context.Context, id uuid.UUID, status models.ReceptionStatus) error {
	return r.receptionQueries.UpdateStatus(ctx, id, status)
}

// GetReceptionProducts получает список товаров приемки
func (r *Repository) GetReceptionProducts(ctx context.Context, receptionID uuid.UUID) ([]*models.Product, error) {
	return r.receptionQueries.GetProducts(ctx, receptionID)
}

// CreateProduct создает новый товар
func (r *Repository) CreateProduct(ctx context.Context, product *models.Product) error {
	return r.productQueries.Create(ctx, product)
}

// CreateProducts создает несколько товаров в одной транзакции
func (r *Repository) CreateProducts(ctx context.Context, products []*models.Product) error {
	return r.productQueries.CreateBatch(ctx, products)
}

// DeleteLastProduct удаляет последний добавленный товар из приемки
func (r *Repository) DeleteLastProduct(ctx context.Context, receptionID uuid.UUID) error {
	return r.productQueries.DeleteLast(ctx, receptionID)
}

// GetProductsByReceptionID получает все товары приемки
func (r *Repository) GetProductsByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]*models.Product, error) {
	return r.productQueries.GetByReceptionID(ctx, receptionID)
}

// CloseReception закрывает приемку
func (r *Repository) CloseReception(ctx context.Context, receptionID uuid.UUID) error {
	fmt.Printf("Closing reception %s with status %s\n", receptionID, models.StatusClose)
	err := r.receptionQueries.UpdateStatus(ctx, receptionID, models.StatusClose)
	if err != nil {
		fmt.Printf("Error closing reception: %v\n", err)
		return err
	}
	fmt.Printf("Reception %s closed successfully\n", receptionID)
	return nil
}

// DeleteProduct удаляет товар
func (r *Repository) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	return r.productQueries.DeleteLast(ctx, productID)
}

// GetLastOpenReception получает последнюю открытую приемку
func (r *Repository) GetLastOpenReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
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

// GetLastProduct получает последний добавленный товар
func (r *Repository) GetLastProduct(ctx context.Context, receptionID uuid.UUID) (*models.Product, error) {
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
	return product, err
}
