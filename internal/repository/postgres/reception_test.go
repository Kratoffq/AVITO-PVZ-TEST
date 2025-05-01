package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/avito/pvz/internal/domain/reception"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReceptionRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewReceptionRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ
	pvzID := uuid.New()
	_, err := db.Exec(`
		INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва');
	`, pvzID)
	require.NoError(t, err)

	tests := []struct {
		name      string
		reception *reception.Reception
		wantErr   bool
	}{
		{
			name: "успешное создание",
			reception: &reception.Reception{
				ID:       uuid.New(),
				DateTime: time.Now(),
				PVZID:    pvzID,
				Status:   reception.StatusInProgress,
			},
			wantErr: false,
		},
		{
			name: "несуществующий ПВЗ",
			reception: &reception.Reception{
				ID:       uuid.New(),
				DateTime: time.Now(),
				PVZID:    uuid.New(),
				Status:   reception.StatusInProgress,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.reception)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Проверяем, что приемка действительно создана
				created, err := repo.GetByID(ctx, tt.reception.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.reception.ID, created.ID)
				assert.Equal(t, tt.reception.Status, created.Status)
				assert.Equal(t, tt.reception.PVZID, created.PVZID)
			}
		})
	}
}

func TestReceptionRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewReceptionRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ и приемку
	pvzID := uuid.New()
	receptionID := uuid.New()
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID, pvzID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		want    *reception.Reception
		wantErr bool
	}{
		{
			name:    "успешное получение",
			id:      receptionID,
			want:    &reception.Reception{ID: receptionID, PVZID: pvzID, Status: reception.StatusInProgress},
			wantErr: false,
		},
		{
			name:    "приемка не найдена",
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
				assert.Equal(t, tt.want.Status, got.Status)
				assert.Equal(t, tt.want.PVZID, got.PVZID)
			}
		})
	}
}

func TestReceptionRepository_GetOpenByPVZID(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewReceptionRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ и приемки
	pvzID := uuid.New()
	receptionID1 := uuid.New()
	receptionID2 := uuid.New()
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID1, pvzID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'close')`, receptionID2, pvzID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		pvzID   uuid.UUID
		want    *reception.Reception
		wantErr bool
	}{
		{
			name:    "успешное получение открытой приемки",
			pvzID:   pvzID,
			want:    &reception.Reception{ID: receptionID1, PVZID: pvzID, Status: reception.StatusInProgress},
			wantErr: false,
		},
		{
			name:    "нет открытых приемок",
			pvzID:   uuid.New(),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetOpenByPVZID(ctx, tt.pvzID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Status, got.Status)
				assert.Equal(t, tt.want.PVZID, got.PVZID)
			}
		})
	}
}

func TestReceptionRepository_Close(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewReceptionRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ и приемку
	pvzID := uuid.New()
	receptionID := uuid.New()
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID, pvzID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "успешное закрытие",
			id:      receptionID,
			wantErr: false,
		},
		{
			name:    "приемка не найдена",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Close(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Проверяем, что приемка действительно закрыта
				closed, err := repo.GetByID(ctx, tt.id)
				assert.NoError(t, err)
				assert.Equal(t, reception.StatusClose, closed.Status)
			}
		})
	}
}

func TestReceptionRepository_List(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewReceptionRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ и приемки
	pvzID := uuid.New()
	_, err := db.Exec(`
		INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва');
	`, pvzID)
	require.NoError(t, err)

	// Создаем несколько приемок
	receptions := []*reception.Reception{
		{
			ID:       uuid.New(),
			DateTime: time.Now(),
			PVZID:    pvzID,
			Status:   reception.StatusInProgress,
		},
		{
			ID:       uuid.New(),
			DateTime: time.Now(),
			PVZID:    pvzID,
			Status:   reception.StatusClose,
		},
		{
			ID:       uuid.New(),
			DateTime: time.Now(),
			PVZID:    pvzID,
			Status:   reception.StatusInProgress,
		},
	}

	for _, r := range receptions {
		err := repo.Create(ctx, r)
		require.NoError(t, err)
	}

	tests := []struct {
		name      string
		offset    int
		limit     int
		wantCount int
		wantErr   bool
	}{
		{
			name:      "получение всех приемок",
			offset:    0,
			limit:     10,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "пагинация с лимитом",
			offset:    0,
			limit:     2,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "пагинация со смещением",
			offset:    2,
			limit:     2,
			wantCount: 1,
			wantErr:   false,
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
				assert.Len(t, got, tt.wantCount)
			}
		})
	}
}

func TestReceptionRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewReceptionRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ и приемку
	pvzID := uuid.New()
	receptionID := uuid.New()
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID, pvzID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		status  reception.Status
		wantErr bool
	}{
		{
			name:    "успешное обновление",
			id:      receptionID,
			status:  reception.StatusClose,
			wantErr: false,
		},
		{
			name:    "приемка не найдена",
			id:      uuid.New(),
			status:  reception.StatusClose,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &reception.Reception{
				ID:     tt.id,
				Status: tt.status,
			}
			err := repo.Update(ctx, r)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Проверяем, что статус действительно обновлен
				updated, err := repo.GetByID(ctx, tt.id)
				assert.NoError(t, err)
				assert.Equal(t, tt.status, updated.Status)
			}
		})
	}
}

func TestReceptionRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewReceptionRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ и приемку
	pvzID := uuid.New()
	receptionID := uuid.New()
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID, pvzID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "успешное удаление",
			id:      receptionID,
			wantErr: false,
		},
		{
			name:    "приемка не найдена",
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
				// Проверяем, что приемка действительно удалена
				_, err := repo.GetByID(ctx, tt.id)
				assert.Error(t, err)
				assert.Equal(t, reception.ErrNotFound, err)
			}
		})
	}
}

func TestReceptionRepository_GetProducts(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewReceptionRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ, приемку и товары
	pvzID := uuid.New()
	receptionID := uuid.New()
	productID1 := uuid.New()
	productID2 := uuid.New()

	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID, pvzID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO products (id, date_time, type, reception_id) VALUES ($1, NOW(), 'electronics', $2)`, productID1, receptionID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO products (id, date_time, type, reception_id) VALUES ($1, NOW(), 'clothing', $2)`, productID2, receptionID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		want    map[uuid.UUID]bool
		wantErr bool
	}{
		{
			name: "успешное получение товаров",
			id:   receptionID,
			want: map[uuid.UUID]bool{
				productID1: true,
				productID2: true,
			},
			wantErr: false,
		},
		{
			name:    "приемка не найдена",
			id:      uuid.New(),
			want:    map[uuid.UUID]bool{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetProducts(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if len(tt.want) > 0 {
					assert.Len(t, got, len(tt.want))
					for _, p := range got {
						assert.True(t, tt.want[p.ID], "unexpected product ID: %s", p.ID)
					}
				} else {
					assert.Empty(t, got)
				}
			}
		})
	}
}

func TestReceptionRepository_GetLastOpen(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewReceptionRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ и приемки
	pvzID := uuid.New()
	receptionID1 := uuid.New()
	receptionID2 := uuid.New()

	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'close')`, receptionID1, pvzID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, 'in_progress')`, receptionID2, pvzID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		pvzID   uuid.UUID
		want    *reception.Reception
		wantErr bool
	}{
		{
			name:    "успешное получение последней открытой приемки",
			pvzID:   pvzID,
			want:    &reception.Reception{ID: receptionID2, PVZID: pvzID, Status: reception.StatusInProgress},
			wantErr: false,
		},
		{
			name:    "нет открытых приемок",
			pvzID:   uuid.New(),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetLastOpen(ctx, tt.pvzID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Status, got.Status)
				assert.Equal(t, tt.want.PVZID, got.PVZID)
			}
		})
	}
}
