package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/avito/pvz/internal/pvz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPVZRepository_Create(t *testing.T) {
	repo := NewMockRepo()
	ctx := context.Background()

	testPVZ := &pvz.PVZ{
		Name:    "Test PVZ",
		Address: "Test Address",
		Schedule: pvz.Schedule{
			OpenTime:  "09:00",
			CloseTime: "21:00",
			Weekend:   false,
		},
		Status:    pvz.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, testPVZ)
	require.NoError(t, err)
	assert.NotZero(t, testPVZ.ID)

	// Проверяем, что PVZ действительно создан
	created, err := repo.GetByID(ctx, testPVZ.ID)
	require.NoError(t, err)
	assert.Equal(t, testPVZ.Name, created.Name)
	assert.Equal(t, testPVZ.Address, created.Address)
	assert.Equal(t, testPVZ.Schedule, created.Schedule)
	assert.Equal(t, testPVZ.Status, created.Status)
}

func TestPVZRepository_GetByID(t *testing.T) {
	repo := NewMockRepo()
	ctx := context.Background()

	// Создаем тестовый PVZ
	testPVZ := &pvz.PVZ{
		Name:    "Test PVZ",
		Address: "Test Address",
		Schedule: pvz.Schedule{
			OpenTime:  "09:00",
			CloseTime: "21:00",
			Weekend:   false,
		},
		Status:    pvz.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, testPVZ)
	require.NoError(t, err)

	// Получаем созданный PVZ
	result, err := repo.GetByID(ctx, testPVZ.ID)
	assert.NoError(t, err)
	assert.Equal(t, testPVZ.ID, result.ID)
	assert.Equal(t, testPVZ.Name, result.Name)
	assert.Equal(t, testPVZ.Address, result.Address)

	// Проверяем случай с несуществующим ID
	_, err = repo.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestPVZRepository_GetLast(t *testing.T) {
	repo := NewMockRepo()
	ctx := context.Background()

	// Проверяем случай с пустым репозиторием
	_, err := repo.GetLast(ctx)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)

	// Создаем тестовый PVZ
	testPVZ := &pvz.PVZ{
		Name:    "Test PVZ",
		Address: "Test Address",
		Schedule: pvz.Schedule{
			OpenTime:  "09:00",
			CloseTime: "21:00",
			Weekend:   false,
		},
		Status:    pvz.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, testPVZ)
	require.NoError(t, err)

	// Получаем последний созданный PVZ
	result, err := repo.GetLast(ctx)
	assert.NoError(t, err)
	assert.Equal(t, testPVZ.ID, result.ID)
	assert.Equal(t, testPVZ.Name, result.Name)
	assert.Equal(t, testPVZ.Address, result.Address)
}

func TestPVZRepository_List(t *testing.T) {
	repo := NewMockRepo()
	ctx := context.Background()

	// Создаем несколько тестовых PVZ
	for i := 0; i < 3; i++ {
		testPVZ := &pvz.PVZ{
			Name:    "Test PVZ",
			Address: "Test Address",
			Schedule: pvz.Schedule{
				OpenTime:  "09:00",
				CloseTime: "21:00",
				Weekend:   false,
			},
			Status:    pvz.StatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(ctx, testPVZ)
		require.NoError(t, err)
	}

	// Получаем список PVZ
	result, err := repo.List(ctx, 0, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestPVZRepository_Update(t *testing.T) {
	repo := NewMockRepo()
	ctx := context.Background()

	// Создаем тестовый PVZ
	testPVZ := &pvz.PVZ{
		Name:    "Test PVZ",
		Address: "Test Address",
		Schedule: pvz.Schedule{
			OpenTime:  "09:00",
			CloseTime: "21:00",
			Weekend:   false,
		},
		Status:    pvz.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, testPVZ)
	require.NoError(t, err)

	// Обновляем PVZ
	testPVZ.Name = "Updated PVZ"
	err = repo.Update(ctx, testPVZ)
	assert.NoError(t, err)

	// Проверяем, что PVZ обновлен
	updated, err := repo.GetByID(ctx, testPVZ.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated PVZ", updated.Name)

	// Проверяем случай с несуществующим PVZ
	nonExistentPVZ := &pvz.PVZ{ID: 999}
	err = repo.Update(ctx, nonExistentPVZ)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestPVZRepository_Delete(t *testing.T) {
	repo := NewMockRepo()
	ctx := context.Background()

	// Создаем тестовый PVZ
	testPVZ := &pvz.PVZ{
		Name:    "Test PVZ",
		Address: "Test Address",
		Schedule: pvz.Schedule{
			OpenTime:  "09:00",
			CloseTime: "21:00",
			Weekend:   false,
		},
		Status:    pvz.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, testPVZ)
	require.NoError(t, err)

	// Удаляем PVZ
	err = repo.Delete(ctx, testPVZ.ID)
	assert.NoError(t, err)

	// Проверяем, что PVZ удален
	_, err = repo.GetByID(ctx, testPVZ.ID)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)

	// Проверяем случай с несуществующим PVZ
	err = repo.Delete(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestPVZRepository_GetWithReceptions(t *testing.T) {
	repo := NewMockRepo()
	ctx := context.Background()

	// Создаем тестовый PVZ
	testPVZ := &pvz.PVZ{
		Name:    "Test PVZ",
		Address: "Test Address",
		Schedule: pvz.Schedule{
			OpenTime:  "09:00",
			CloseTime: "21:00",
			Weekend:   false,
		},
		Status:    pvz.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, testPVZ)
	require.NoError(t, err)

	// Получаем PVZ с приемами
	pvz, receptions, err := repo.GetWithReceptions(ctx, testPVZ.ID)
	assert.NoError(t, err)
	assert.Equal(t, testPVZ.ID, pvz.ID)
	assert.Empty(t, receptions)

	// Проверяем случай с несуществующим PVZ
	_, _, err = repo.GetWithReceptions(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}
