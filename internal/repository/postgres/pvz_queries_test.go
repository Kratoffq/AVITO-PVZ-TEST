package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/avito/pvz/internal/models"
	"github.com/avito/pvz/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDBURL = "postgres://postgres:postgres@localhost:5434/pvz_test?sslmode=disable"

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", testDBURL)
	require.NoError(t, err)

	// Очищаем таблицы перед тестом
	_, err = db.Exec("TRUNCATE pvzs, receptions, products, users RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	return db
}

func TestPVZQueries_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	queries := NewPVZQueries(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		pvz     *models.PVZ
		wantErr bool
	}{
		{
			name: "successful creation",
			pvz: &models.PVZ{
				ID:               uuid.New(),
				RegistrationDate: time.Now(),
				City:             "Москва",
			},
			wantErr: false,
		},
		{
			name: "invalid city",
			pvz: &models.PVZ{
				ID:               uuid.New(),
				RegistrationDate: time.Now(),
				City:             "Новосибирск", // Невалидный город
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := queries.Create(ctx, tt.pvz)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Проверяем, что ПВЗ создался
				created, err := queries.GetByID(ctx, tt.pvz.ID)
				require.NoError(t, err)
				assert.Equal(t, tt.pvz.ID, created.ID)
				assert.Equal(t, tt.pvz.City, created.City)
			}
		})
	}
}

func TestPVZQueries_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	queries := NewPVZQueries(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	err := queries.Create(ctx, pvz)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		want    *models.PVZ
		wantErr bool
	}{
		{
			name: "existing PVZ",
			id:   pvz.ID,
			want: pvz,
		},
		{
			name: "non-existing PVZ",
			id:   uuid.New(),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := queries.GetByID(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.want == nil {
					assert.Nil(t, got)
				} else {
					assert.Equal(t, tt.want.ID, got.ID)
					assert.Equal(t, tt.want.City, got.City)
				}
			}
		})
	}
}

func TestPVZQueries_GetWithReceptions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	queries := NewPVZQueries(db)
	ctx := context.Background()

	// Создаем тестовые данные
	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	err := queries.Create(ctx, pvz)
	require.NoError(t, err)

	// Создаем приемку
	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvz.ID,
		Status:   models.StatusInProgress,
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO receptions (id, date_time, pvz_id, status)
		VALUES ($1, $2, $3, $4)
	`, reception.ID, reception.DateTime, reception.PVZID, reception.Status)
	require.NoError(t, err)

	// Создаем товар
	product := &models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        models.TypeElectronics,
		ReceptionID: reception.ID,
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO products (id, date_time, type, reception_id)
		VALUES ($1, $2, $3, $4)
	`, product.ID, product.DateTime, product.Type, product.ReceptionID)
	require.NoError(t, err)

	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		page      int
		limit     int
		want      []*repository.PVZWithReceptions
		wantErr   bool
	}{
		{
			name:      "get PVZ with receptions",
			startDate: time.Now().Add(-24 * time.Hour),
			endDate:   time.Now().Add(24 * time.Hour),
			page:      1,
			limit:     10,
			want: []*repository.PVZWithReceptions{
				{
					PVZ: pvz,
					Receptions: []*repository.ReceptionWithProducts{
						{
							Reception: reception,
							Products:  []*models.Product{product},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := queries.GetWithReceptions(ctx, tt.startDate, tt.endDate, tt.page, tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.Len(t, got, len(tt.want))

				// Проверяем ПВЗ
				assert.Equal(t, tt.want[0].PVZ.ID, got[0].PVZ.ID)
				assert.Equal(t, tt.want[0].PVZ.City, got[0].PVZ.City)

				// Проверяем приемки
				require.Len(t, got[0].Receptions, len(tt.want[0].Receptions))
				assert.Equal(t, tt.want[0].Receptions[0].Reception.ID, got[0].Receptions[0].Reception.ID)
				assert.Equal(t, tt.want[0].Receptions[0].Reception.Status, got[0].Receptions[0].Reception.Status)

				// Проверяем товары
				require.Len(t, got[0].Receptions[0].Products, len(tt.want[0].Receptions[0].Products))
				assert.Equal(t, tt.want[0].Receptions[0].Products[0].ID, got[0].Receptions[0].Products[0].ID)
				assert.Equal(t, tt.want[0].Receptions[0].Products[0].Type, got[0].Receptions[0].Products[0].Type)
			}
		})
	}
}

func TestPVZQueries_GetAllPVZs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	queries := NewPVZQueries(db)
	ctx := context.Background()

	// Создаем тестовые ПВЗ
	pvzs := []*models.PVZ{
		{
			ID:               uuid.New(),
			RegistrationDate: time.Now(),
			City:             "Москва",
		},
		{
			ID:               uuid.New(),
			RegistrationDate: time.Now(),
			City:             "Москва",
		},
	}

	for _, pvz := range pvzs {
		err := queries.Create(ctx, pvz)
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		want    []*models.PVZ
		wantErr bool
	}{
		{
			name: "get all PVZs",
			want: pvzs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := queries.GetAllPVZs(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.Len(t, got, len(tt.want))

				// Проверяем, что все ПВЗ найдены
				found := make(map[uuid.UUID]bool)
				for _, pvz := range got {
					found[pvz.ID] = true
				}
				for _, pvz := range tt.want {
					assert.True(t, found[pvz.ID], "PVZ %s not found", pvz.ID)
				}
			}
		})
	}
}
