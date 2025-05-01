package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/avito/pvz/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReceptionQueries_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	queries := NewReceptionQueries(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	pvzQueries := NewPVZQueries(db)
	err := pvzQueries.Create(ctx, pvz)
	require.NoError(t, err)

	tests := []struct {
		name      string
		reception *models.Reception
		wantErr   bool
	}{
		{
			name: "successful creation",
			reception: &models.Reception{
				ID:       uuid.New(),
				DateTime: time.Now(),
				PVZID:    pvz.ID,
				Status:   models.StatusInProgress,
			},
			wantErr: false,
		},
		{
			name: "invalid PVZ ID",
			reception: &models.Reception{
				ID:       uuid.New(),
				DateTime: time.Now(),
				PVZID:    uuid.New(), // Несуществующий ПВЗ
				Status:   models.StatusInProgress,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := queries.Create(ctx, tt.reception)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Проверяем, что приемка создалась
				created, err := queries.GetByID(ctx, tt.reception.ID)
				require.NoError(t, err)
				assert.Equal(t, tt.reception.ID, created.ID)
				assert.Equal(t, tt.reception.PVZID, created.PVZID)
				assert.Equal(t, tt.reception.Status, created.Status)
			}
		})
	}
}

func TestReceptionQueries_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	queries := NewReceptionQueries(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	pvzQueries := NewPVZQueries(db)
	err := pvzQueries.Create(ctx, pvz)
	require.NoError(t, err)

	// Создаем тестовую приемку
	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvz.ID,
		Status:   models.StatusInProgress,
	}
	err = queries.Create(ctx, reception)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		want    *models.Reception
		wantErr bool
	}{
		{
			name: "existing reception",
			id:   reception.ID,
			want: reception,
		},
		{
			name: "non-existing reception",
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
					assert.Equal(t, tt.want.PVZID, got.PVZID)
					assert.Equal(t, tt.want.Status, got.Status)
				}
			}
		})
	}
}

func TestReceptionQueries_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	queries := NewReceptionQueries(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	pvzQueries := NewPVZQueries(db)
	err := pvzQueries.Create(ctx, pvz)
	require.NoError(t, err)

	// Создаем тестовую приемку
	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvz.ID,
		Status:   models.StatusInProgress,
	}
	err = queries.Create(ctx, reception)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		status  models.ReceptionStatus
		wantErr bool
	}{
		{
			name:    "successful update",
			id:      reception.ID,
			status:  models.StatusClose,
			wantErr: false,
		},
		{
			name:    "non-existing reception",
			id:      uuid.New(),
			status:  models.StatusClose,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := queries.UpdateStatus(ctx, tt.id, tt.status)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Проверяем, что статус обновился
				updated, err := queries.GetByID(ctx, tt.id)
				require.NoError(t, err)
				assert.Equal(t, tt.status, updated.Status)
			}
		})
	}
}

func TestReceptionQueries_GetProducts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	queries := NewReceptionQueries(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	pvzQueries := NewPVZQueries(db)
	err := pvzQueries.Create(ctx, pvz)
	require.NoError(t, err)

	// Создаем тестовую приемку
	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvz.ID,
		Status:   models.StatusInProgress,
	}
	err = queries.Create(ctx, reception)
	require.NoError(t, err)

	// Создаем тестовые товары
	products := []*models.Product{
		{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        models.TypeElectronics,
			ReceptionID: reception.ID,
		},
		{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        models.TypeClothing,
			ReceptionID: reception.ID,
		},
	}

	for _, product := range products {
		_, err = db.ExecContext(ctx, `
			INSERT INTO products (id, date_time, type, reception_id)
			VALUES ($1, $2, $3, $4)
		`, product.ID, product.DateTime, product.Type, product.ReceptionID)
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		receptionID uuid.UUID
		want        []*models.Product
		wantErr     bool
	}{
		{
			name:        "get products for existing reception",
			receptionID: reception.ID,
			want:        products,
		},
		{
			name:        "get products for non-existing reception",
			receptionID: uuid.New(),
			want:        []*models.Product{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := queries.GetProducts(ctx, tt.receptionID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.Len(t, got, len(tt.want))

				// Проверяем, что все товары найдены
				found := make(map[uuid.UUID]bool)
				for _, product := range got {
					found[product.ID] = true
				}
				for _, product := range tt.want {
					assert.True(t, found[product.ID], "Product %s not found", product.ID)
				}
			}
		})
	}
}
