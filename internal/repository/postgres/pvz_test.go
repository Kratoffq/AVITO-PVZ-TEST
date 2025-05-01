package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/avito/pvz/internal/domain/pvz"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPVZRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewPVZRepository(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		pvz     *pvz.PVZ
		wantErr bool
	}{
		{
			name: "успешное создание",
			pvz: &pvz.PVZ{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				City:      "Москва",
			},
			wantErr: false,
		},
		{
			name: "пустой город",
			pvz: &pvz.PVZ{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				City:      "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.pvz)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Проверяем, что ПВЗ действительно создан
				created, err := repo.GetByID(ctx, tt.pvz.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.pvz.ID, created.ID)
				assert.Equal(t, tt.pvz.City, created.City)
			}
		})
	}
}

func TestPVZRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewPVZRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ
	pvzID := uuid.New()
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		want    *pvz.PVZ
		wantErr bool
	}{
		{
			name: "успешное получение",
			id:   pvzID,
			want: &pvz.PVZ{
				ID:   pvzID,
				City: "Москва",
			},
			wantErr: false,
		},
		{
			name:    "ПВЗ не найден",
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
				assert.Equal(t, tt.want.City, got.City)
			}
		})
	}
}

func TestPVZRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewPVZRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ
	pvzID := uuid.New()
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		pvz     *pvz.PVZ
		wantErr bool
	}{
		{
			name: "успешное обновление",
			pvz: &pvz.PVZ{
				ID:   pvzID,
				City: "Санкт-Петербург",
			},
			wantErr: false,
		},
		{
			name: "ПВЗ не существует",
			pvz: &pvz.PVZ{
				ID:   uuid.New(),
				City: "Санкт-Петербург",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(ctx, tt.pvz)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Проверяем, что ПВЗ действительно обновлен
				updated, err := repo.GetByID(ctx, tt.pvz.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.pvz.City, updated.City)
			}
		})
	}
}

func TestPVZRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewPVZRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ
	pvzID := uuid.New()
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "успешное удаление",
			id:      pvzID,
			wantErr: false,
		},
		{
			name:    "ПВЗ не существует",
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
				// Проверяем, что ПВЗ действительно удален
				_, err := repo.GetByID(ctx, tt.id)
				assert.Error(t, err)
			}
		})
	}
}

func TestPVZRepository_List(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewPVZRepository(db)
	ctx := context.Background()

	// Создаем несколько тестовых ПВЗ
	pvzs := []*pvz.PVZ{
		{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			City:      "Москва",
		},
		{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			City:      "Санкт-Петербург",
		},
		{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			City:      "Казань",
		},
	}

	for _, p := range pvzs {
		_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, $2, $3)`, p.ID, p.CreatedAt, p.City)
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
			name:    "получение всех ПВЗ",
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

func TestPVZRepository_GetByCity(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewPVZRepository(db)
	ctx := context.Background()

	// Создаем тестовые ПВЗ
	cities := []string{"Москва", "Москва", "Санкт-Петербург"}
	for _, city := range cities {
		_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), $2)`, uuid.New(), city)
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		city    string
		want    *pvz.PVZ
		wantErr bool
	}{
		{
			name:    "успешное получение",
			city:    "Москва",
			want:    &pvz.PVZ{City: "Москва"},
			wantErr: false,
		},
		{
			name:    "город не найден",
			city:    "Новосибирск",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByCity(ctx, tt.city)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.city, got.City)
			}
		})
	}
}

func TestPVZRepository_GetWithReceptions(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewPVZRepository(db)
	ctx := context.Background()

	// Создаем тестовый ПВЗ и приемки
	pvzID := uuid.New()
	_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), 'Москва')`, pvzID)
	require.NoError(t, err)

	receptionID1 := uuid.New()
	receptionID2 := uuid.New()

	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, $3)`,
		receptionID1, pvzID, reception.StatusInProgress)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, NOW(), $2, $3)`,
		receptionID2, pvzID, reception.StatusInProgress)
	require.NoError(t, err)

	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		page      int
		limit     int
		wantLen   int
		wantErr   bool
	}{
		{
			name:      "успешное получение",
			startDate: time.Now().Add(-24 * time.Hour),
			endDate:   time.Now().Add(24 * time.Hour),
			page:      1,
			limit:     10,
			wantLen:   1,
			wantErr:   false,
		},
		{
			name:      "нет приемок в период",
			startDate: time.Now().Add(24 * time.Hour),
			endDate:   time.Now().Add(48 * time.Hour),
			page:      1,
			limit:     10,
			wantLen:   0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetWithReceptions(ctx, tt.startDate, tt.endDate, tt.page, tt.limit)
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

func TestPVZRepository_GetAll(t *testing.T) {
	db := SetupTestDB(t)
	repo := NewPVZRepository(db)
	ctx := context.Background()

	// Создаем несколько тестовых ПВЗ
	cities := []string{"Москва", "Санкт-Петербург", "Казань"}
	for _, city := range cities {
		_, err := db.Exec(`INSERT INTO pvzs (id, created_at, city) VALUES ($1, NOW(), $2)`, uuid.New(), city)
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		wantLen int
		wantErr bool
	}{
		{
			name:    "успешное получение всех ПВЗ",
			wantLen: 3,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetAll(ctx)
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
