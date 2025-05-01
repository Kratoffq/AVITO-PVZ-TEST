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

func TestProductQueries_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	queries := NewProductQueries(db)
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
	receptionQueries := NewReceptionQueries(db)
	err = receptionQueries.Create(ctx, reception)
	require.NoError(t, err)

	tests := []struct {
		name    string
		product *models.Product
		wantErr bool
	}{
		{
			name: "successful creation",
			product: &models.Product{
				ID:          uuid.New(),
				DateTime:    time.Now(),
				Type:        models.TypeElectronics,
				ReceptionID: reception.ID,
			},
			wantErr: false,
		},
		{
			name: "invalid reception ID",
			product: &models.Product{
				ID:          uuid.New(),
				DateTime:    time.Now(),
				Type:        models.TypeElectronics,
				ReceptionID: uuid.New(), // Несуществующая приемка
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := queries.Create(ctx, tt.product)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Проверяем, что товар создался
				products, err := queries.GetByReceptionID(ctx, tt.product.ReceptionID)
				require.NoError(t, err)
				require.Len(t, products, 1)
				assert.Equal(t, tt.product.ID, products[0].ID)
				assert.Equal(t, tt.product.Type, products[0].Type)
			}
		})
	}
}

func TestProductQueries_CreateBatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	queries := NewProductQueries(db)
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
	receptionQueries := NewReceptionQueries(db)
	err = receptionQueries.Create(ctx, reception)
	require.NoError(t, err)

	tests := []struct {
		name     string
		products []*models.Product
		wantErr  bool
	}{
		{
			name: "successful batch creation",
			products: []*models.Product{
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
			},
			wantErr: false,
		},
		{
			name: "invalid reception ID in batch",
			products: []*models.Product{
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
					ReceptionID: uuid.New(), // Несуществующая приемка
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := queries.CreateBatch(ctx, tt.products)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Проверяем, что все товары создались
				products, err := queries.GetByReceptionID(ctx, reception.ID)
				require.NoError(t, err)
				require.Len(t, products, len(tt.products))

				// Проверяем, что все товары найдены
				found := make(map[uuid.UUID]bool)
				for _, product := range products {
					found[product.ID] = true
				}
				for _, product := range tt.products {
					assert.True(t, found[product.ID], "Product %s not found", product.ID)
				}
			}
		})
	}
}

func TestProductQueries_DeleteLast(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	queries := NewProductQueries(db)
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
	receptionQueries := NewReceptionQueries(db)
	err = receptionQueries.Create(ctx, reception)
	require.NoError(t, err)

	// Создаем тестовые товары
	products := []*models.Product{
		{
			ID:          uuid.New(),
			DateTime:    time.Now().Add(-time.Hour),
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

	err = queries.CreateBatch(ctx, products)
	require.NoError(t, err)

	tests := []struct {
		name        string
		receptionID uuid.UUID
		wantErr     bool
	}{
		{
			name:        "successful deletion",
			receptionID: reception.ID,
			wantErr:     false,
		},
		{
			name:        "non-existing reception",
			receptionID: uuid.New(),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := queries.DeleteLast(ctx, tt.receptionID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Проверяем, что последний товар удалился
				remaining, err := queries.GetByReceptionID(ctx, tt.receptionID)
				require.NoError(t, err)
				require.Len(t, remaining, 1)
				assert.Equal(t, products[0].ID, remaining[0].ID)
			}
		})
	}
}

func TestProductQueries_GetByReceptionID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	queries := NewProductQueries(db)
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
	receptionQueries := NewReceptionQueries(db)
	err = receptionQueries.Create(ctx, reception)
	require.NoError(t, err)

	// Создаем тестовые товары
	products := []*models.Product{
		{
			ID:          uuid.New(),
			DateTime:    time.Now().Add(-time.Hour),
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

	err = queries.CreateBatch(ctx, products)
	require.NoError(t, err)

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
			got, err := queries.GetByReceptionID(ctx, tt.receptionID)
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
