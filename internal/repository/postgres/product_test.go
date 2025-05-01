package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/avito/pvz/internal/repository/postgres/queries"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ и приемку
	pvzID := uuid.New()
	receptionID := uuid.New()

	// Разделяем SQL-запросы
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID, pvzID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		product *product.Product
		wantErr bool
	}{
		{
			name: "успешное создание",
			product: &product.Product{
				ID:          uuid.New(),
				DateTime:    time.Now(),
				Type:        product.TypeElectronics,
				ReceptionID: receptionID,
			},
			wantErr: false,
		},
		{
			name: "неверный тип товара",
			product: &product.Product{
				ID:          uuid.New(),
				DateTime:    time.Now(),
				Type:        "invalid_type",
				ReceptionID: receptionID,
			},
			wantErr: true,
		},
		{
			name: "несуществующая приемка",
			product: &product.Product{
				ID:          uuid.New(),
				DateTime:    time.Now(),
				Type:        product.TypeElectronics,
				ReceptionID: uuid.New(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Получаем SQL-запрос и его параметры
			query, args, err := queries.CreateProduct(tt.product.ID, tt.product.DateTime, string(tt.product.Type), tt.product.ReceptionID)
			require.NoError(t, err)
			t.Logf("SQL Query: %s, Args: %v", query, args)

			err = repo.Create(ctx, tt.product)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Проверяем, что товар действительно создан
				created, err := repo.GetByID(ctx, tt.product.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.product.ID, created.ID)
				assert.Equal(t, tt.product.Type, created.Type)
				assert.Equal(t, tt.product.ReceptionID, created.ReceptionID)
			}
		})
	}
}

func TestProductRepository_CreateBatch(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ и приемку
	pvzID := uuid.New()
	receptionID := uuid.New()

	// Разделяем SQL-запросы
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID, pvzID)
	require.NoError(t, err)

	tests := []struct {
		name     string
		products []*product.Product
		wantErr  bool
	}{
		{
			name: "успешное создание нескольких товаров",
			products: []*product.Product{
				{
					ID:          uuid.New(),
					DateTime:    time.Now(),
					Type:        product.TypeElectronics,
					ReceptionID: receptionID,
				},
				{
					ID:          uuid.New(),
					DateTime:    time.Now(),
					Type:        product.TypeClothing,
					ReceptionID: receptionID,
				},
			},
			wantErr: false,
		},
		{
			name: "ошибка при создании одного из товаров",
			products: []*product.Product{
				{
					ID:          uuid.New(),
					DateTime:    time.Now(),
					Type:        product.TypeElectronics,
					ReceptionID: receptionID,
				},
				{
					ID:          uuid.New(),
					DateTime:    time.Now(),
					Type:        "invalid_type",
					ReceptionID: receptionID,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateBatch(ctx, tt.products)
			if tt.wantErr {
				assert.Error(t, err)
				// Проверяем, что ни один товар не был создан (транзакция откатилась)
				for _, p := range tt.products {
					_, err := repo.GetByID(ctx, p.ID)
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
				// Проверяем, что все товары созданы
				for _, p := range tt.products {
					created, err := repo.GetByID(ctx, p.ID)
					assert.NoError(t, err)
					assert.Equal(t, p.ID, created.ID)
					assert.Equal(t, p.Type, created.Type)
					assert.Equal(t, p.ReceptionID, created.ReceptionID)
				}
			}
		})
	}
}

func TestProductRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ, приемку и товар
	pvzID := uuid.New()
	receptionID := uuid.New()
	productID := uuid.New()

	// Разделяем SQL-запросы
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID, pvzID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO products (id, date_time, type, reception_id) VALUES ($1, NOW(), $2, $3)`, productID, product.TypeElectronics, receptionID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		want    *product.Product
		wantErr bool
	}{
		{
			name: "успешное получение",
			id:   productID,
			want: &product.Product{
				ID:          productID,
				Type:        product.TypeElectronics,
				ReceptionID: receptionID,
			},
			wantErr: false,
		},
		{
			name:    "товар не найден",
			id:      uuid.New(),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Type, got.Type)
				assert.Equal(t, tt.want.ReceptionID, got.ReceptionID)
			}
		})
	}
}

func TestProductRepository_GetByReceptionID(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ и приемку
	pvzID := uuid.New()
	receptionID := uuid.New()

	// Разделяем SQL-запросы
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID, pvzID)
	require.NoError(t, err)

	// Создаем несколько товаров
	products := []*product.Product{
		{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        product.TypeElectronics,
			ReceptionID: receptionID,
		},
		{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        product.TypeClothing,
			ReceptionID: receptionID,
		},
	}

	for _, p := range products {
		err := repo.Create(ctx, p)
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		receptionID uuid.UUID
		wantCount   int
		wantErr     bool
	}{
		{
			name:        "успешное получение товаров",
			receptionID: receptionID,
			wantCount:   2,
			wantErr:     false,
		},
		{
			name:        "нет товаров в приемке",
			receptionID: uuid.New(),
			wantCount:   0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByReceptionID(ctx, tt.receptionID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantCount)
			}
		})
	}
}

func TestProductRepository_DeleteLast(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ и приемку
	pvzID := uuid.New()
	receptionID := uuid.New()

	// Разделяем SQL-запросы
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID, pvzID)
	require.NoError(t, err)

	// Создаем несколько товаров
	products := []*product.Product{
		{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        product.TypeElectronics,
			ReceptionID: receptionID,
		},
		{
			ID:          uuid.New(),
			DateTime:    time.Now().Add(time.Hour),
			Type:        product.TypeClothing,
			ReceptionID: receptionID,
		},
	}

	for _, p := range products {
		err := repo.Create(ctx, p)
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		receptionID uuid.UUID
		wantErr     bool
	}{
		{
			name:        "успешное удаление последнего товара",
			receptionID: receptionID,
			wantErr:     false,
		},
		{
			name:        "нет товаров в приемке",
			receptionID: uuid.New(),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.DeleteLast(ctx, tt.receptionID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Проверяем, что остался только первый товар
				remaining, err := repo.GetByReceptionID(ctx, tt.receptionID)
				assert.NoError(t, err)
				assert.Len(t, remaining, 1)
				assert.Equal(t, products[0].ID, remaining[0].ID)
			}
		})
	}
}

func TestProductRepository_List(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ и приемку
	pvzID := uuid.New()
	receptionID := uuid.New()

	// Разделяем SQL-запросы
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID, pvzID)
	require.NoError(t, err)

	// Создаем несколько товаров
	products := []*product.Product{
		{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        product.TypeElectronics,
			ReceptionID: receptionID,
		},
		{
			ID:          uuid.New(),
			DateTime:    time.Now().Add(time.Hour),
			Type:        product.TypeClothing,
			ReceptionID: receptionID,
		},
		{
			ID:          uuid.New(),
			DateTime:    time.Now().Add(2 * time.Hour),
			Type:        product.TypeFood,
			ReceptionID: receptionID,
		},
	}

	for _, p := range products {
		err := repo.Create(ctx, p)
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		limit   int
		offset  int
		wantLen int
		wantErr bool
	}{
		{
			name:    "получение всех товаров",
			limit:   10,
			offset:  0,
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "пагинация",
			limit:   2,
			offset:  0,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "смещение",
			limit:   2,
			offset:  1,
			wantLen: 2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.List(ctx, tt.offset, tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

func TestProductRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ, приемку и товар
	pvzID := uuid.New()
	receptionID := uuid.New()
	productID := uuid.New()

	// Разделяем SQL-запросы
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID, pvzID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO products (id, date_time, type, reception_id) VALUES ($1, NOW(), $2, $3)`, productID, product.TypeElectronics, receptionID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "успешное удаление",
			id:      productID,
			wantErr: false,
		},
		{
			name:    "товар не существует",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Проверяем, что товар действительно удален
				_, err := repo.GetByID(ctx, tt.id)
				assert.Error(t, err)
			}
		})
	}
}

func TestProductRepository_GetLast(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ и приемку
	pvzID := uuid.New()
	receptionID := uuid.New()

	// Разделяем SQL-запросы
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID, pvzID)
	require.NoError(t, err)

	// Создаем несколько товаров
	products := []*product.Product{
		{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        product.TypeElectronics,
			ReceptionID: receptionID,
		},
		{
			ID:          uuid.New(),
			DateTime:    time.Now().Add(time.Hour),
			Type:        product.TypeClothing,
			ReceptionID: receptionID,
		},
	}

	for _, p := range products {
		err := repo.Create(ctx, p)
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		receptionID uuid.UUID
		want        *product.Product
		wantErr     bool
	}{
		{
			name:        "успешное получение последнего товара",
			receptionID: receptionID,
			want:        products[1],
			wantErr:     false,
		},
		{
			name:        "нет товаров в приемке",
			receptionID: uuid.New(),
			want:        nil,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetLast(ctx, tt.receptionID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if tt.want == nil {
					assert.Nil(t, got)
				} else {
					assert.NotNil(t, got)
					assert.Equal(t, tt.want.ID, got.ID)
					assert.Equal(t, tt.want.Type, got.Type)
					assert.Equal(t, tt.want.ReceptionID, got.ReceptionID)
				}
			}
		})
	}
}

func TestProductRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ, приемку и товар
	pvzID := uuid.New()
	receptionID := uuid.New()
	productID := uuid.New()

	// Разделяем SQL-запросы
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID, pvzID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO products (id, date_time, type, reception_id) VALUES ($1, NOW(), $2, $3)`, productID, product.TypeElectronics, receptionID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		product *product.Product
		wantErr bool
	}{
		{
			name: "успешное обновление",
			product: &product.Product{
				ID:          productID,
				DateTime:    time.Now().Add(time.Hour),
				Type:        product.TypeClothing,
				ReceptionID: receptionID,
			},
			wantErr: false,
		},
		{
			name: "товар не существует",
			product: &product.Product{
				ID:          uuid.New(),
				DateTime:    time.Now(),
				Type:        product.TypeElectronics,
				ReceptionID: receptionID,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(ctx, tt.product)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Проверяем, что товар действительно обновлен
				updated, err := repo.GetByID(ctx, tt.product.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.product.Type, updated.Type)
				assert.Equal(t, tt.product.DateTime.Unix(), updated.DateTime.Unix())
			}
		})
	}
}
