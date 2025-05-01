package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/avito/pvz/internal/pvz"
	"github.com/avito/pvz/internal/pvz/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errMock = errors.New("mock error")

func TestPVZService_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := repository.NewMockRepo()
		txManager := repository.NewMockTxManager()
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		pvz := &pvz.PVZ{
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

		err := service.Create(ctx, pvz)
		require.NoError(t, err)
		assert.NotZero(t, pvz.ID)
	})

	t.Run("begin tx error", func(t *testing.T) {
		repo := repository.NewMockRepo()
		txManager := repository.NewMockTxManager()
		txManager.BeginTxError = errMock
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		err := service.Create(ctx, &pvz.PVZ{})
		assert.ErrorIs(t, err, errMock)
	})

	t.Run("create error with rollback", func(t *testing.T) {
		repo := repository.NewMockRepo()
		repo.CreateError = errMock
		txManager := repository.NewMockTxManager()
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		err := service.Create(ctx, &pvz.PVZ{})
		assert.ErrorIs(t, err, errMock)
	})

	t.Run("create error with rollback error", func(t *testing.T) {
		repo := repository.NewMockRepo()
		repo.CreateError = errMock
		txManager := repository.NewMockTxManager()
		txManager.RollbackTxError = errors.New("rollback error")
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		err := service.Create(ctx, &pvz.PVZ{})
		assert.ErrorIs(t, err, txManager.RollbackTxError)
	})
}

func TestPVZService_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := repository.NewMockRepo()
		txManager := repository.NewMockTxManager()
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		// Сначала создаем PVZ
		pvz := &pvz.PVZ{
			Name:    "Test PVZ",
			Address: "Test Address",
		}
		err := service.Create(ctx, pvz)
		require.NoError(t, err)

		// Обновляем PVZ
		pvz.Name = "Updated PVZ"
		err = service.Update(ctx, pvz)
		require.NoError(t, err)

		// Проверяем, что PVZ обновлен
		updated, err := service.GetByID(ctx, pvz.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated PVZ", updated.Name)
	})

	t.Run("begin tx error", func(t *testing.T) {
		repo := repository.NewMockRepo()
		txManager := repository.NewMockTxManager()
		txManager.BeginTxError = errMock
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		err := service.Update(ctx, &pvz.PVZ{})
		assert.ErrorIs(t, err, errMock)
	})

	t.Run("update error with rollback", func(t *testing.T) {
		repo := repository.NewMockRepo()
		repo.UpdateError = errMock
		txManager := repository.NewMockTxManager()
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		err := service.Update(ctx, &pvz.PVZ{})
		assert.ErrorIs(t, err, errMock)
	})

	t.Run("update error with commit error", func(t *testing.T) {
		repo := repository.NewMockRepo()
		txManager := repository.NewMockTxManager()
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		// Сначала создаем PVZ
		pvz := &pvz.PVZ{
			Name:    "Test PVZ",
			Address: "Test Address",
		}
		txManager.CommitTxError = nil // Сбрасываем ошибку для создания
		err := service.Create(ctx, pvz)
		require.NoError(t, err)

		// Устанавливаем ошибку для обновления
		txManager.CommitTxError = errMock
		err = service.Update(ctx, pvz)
		assert.ErrorIs(t, err, errMock)
	})
}

func TestPVZService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := repository.NewMockRepo()
		txManager := repository.NewMockTxManager()
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		// Сначала создаем PVZ
		pvz := &pvz.PVZ{
			Name:    "Test PVZ",
			Address: "Test Address",
		}
		err := service.Create(ctx, pvz)
		require.NoError(t, err)

		// Удаляем PVZ
		err = service.Delete(ctx, pvz.ID)
		require.NoError(t, err)

		// Проверяем, что PVZ удален
		_, err = service.GetByID(ctx, pvz.ID)
		assert.Error(t, err)
	})

	t.Run("begin tx error", func(t *testing.T) {
		repo := repository.NewMockRepo()
		txManager := repository.NewMockTxManager()
		txManager.BeginTxError = errMock
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		err := service.Delete(ctx, 1)
		assert.ErrorIs(t, err, errMock)
	})

	t.Run("delete error with rollback", func(t *testing.T) {
		repo := repository.NewMockRepo()
		repo.DeleteError = errMock
		txManager := repository.NewMockTxManager()
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		err := service.Delete(ctx, 1)
		assert.ErrorIs(t, err, errMock)
	})

	t.Run("delete error with commit error", func(t *testing.T) {
		repo := repository.NewMockRepo()
		txManager := repository.NewMockTxManager()
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		// Сначала создаем PVZ
		pvz := &pvz.PVZ{
			Name:    "Test PVZ",
			Address: "Test Address",
		}
		txManager.CommitTxError = nil // Сбрасываем ошибку для создания
		err := service.Create(ctx, pvz)
		require.NoError(t, err)

		// Устанавливаем ошибку для удаления
		txManager.CommitTxError = errMock
		err = service.Delete(ctx, pvz.ID)
		assert.ErrorIs(t, err, errMock)
	})
}

func TestPVZService_GetByID(t *testing.T) {
	repo := repository.NewMockRepo()
	txManager := repository.NewMockTxManager()
	service := NewPVZService(repo, txManager)
	ctx := context.Background()

	// Создаем тестовый PVZ
	testPVZ := &pvz.PVZ{
		Name:    "Test PVZ",
		Address: "Test Address",
	}

	err := service.Create(ctx, testPVZ)
	require.NoError(t, err)

	// Получаем PVZ по ID
	found, err := service.GetByID(ctx, testPVZ.ID)
	require.NoError(t, err)
	assert.Equal(t, testPVZ.Name, found.Name)
	assert.Equal(t, testPVZ.Address, found.Address)

	// Проверяем случай с несуществующим PVZ
	_, err = service.GetByID(ctx, 999)
	assert.Error(t, err)
}

func TestPVZService_GetLast(t *testing.T) {
	repo := repository.NewMockRepo()
	txManager := repository.NewMockTxManager()
	service := NewPVZService(repo, txManager)
	ctx := context.Background()

	// Проверяем случай с пустым репозиторием
	_, err := service.GetLast(ctx)
	assert.Error(t, err)

	// Создаем тестовый PVZ
	testPVZ := &pvz.PVZ{
		Name:    "Test PVZ",
		Address: "Test Address",
	}

	err = service.Create(ctx, testPVZ)
	require.NoError(t, err)

	// Получаем последний PVZ
	last, err := service.GetLast(ctx)
	require.NoError(t, err)
	assert.Equal(t, testPVZ.Name, last.Name)
	assert.Equal(t, testPVZ.Address, last.Address)
}

func TestPVZService_List(t *testing.T) {
	repo := repository.NewMockRepo()
	txManager := repository.NewMockTxManager()
	service := NewPVZService(repo, txManager)
	ctx := context.Background()

	// Создаем несколько тестовых PVZ
	for i := 0; i < 3; i++ {
		err := service.Create(ctx, &pvz.PVZ{
			Name:    "Test PVZ",
			Address: "Test Address",
		})
		require.NoError(t, err)
	}

	// Получаем список PVZ
	list, err := service.List(ctx, 0, 10)
	require.NoError(t, err)
	assert.Len(t, list, 3)
}

func TestPVZService_GetWithReceptions(t *testing.T) {
	repo := repository.NewMockRepo()
	txManager := repository.NewMockTxManager()
	service := NewPVZService(repo, txManager)
	ctx := context.Background()

	// Создаем тестовый PVZ
	testPVZ := &pvz.PVZ{
		Name:    "Test PVZ",
		Address: "Test Address",
	}

	err := service.Create(ctx, testPVZ)
	require.NoError(t, err)

	// Получаем PVZ с приемами товаров
	pvz, receptions, err := service.GetWithReceptions(ctx, testPVZ.ID)
	require.NoError(t, err)
	assert.Equal(t, testPVZ.Name, pvz.Name)
	assert.Equal(t, testPVZ.Address, pvz.Address)
	assert.Empty(t, receptions)

	// Проверяем случай с несуществующим PVZ
	_, _, err = service.GetWithReceptions(ctx, 999)
	assert.Error(t, err)
}

func TestPVZService_GetPvzList(t *testing.T) {
	t.Run("success with default pagination", func(t *testing.T) {
		repo := repository.NewMockRepo()
		txManager := repository.NewMockTxManager()
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		// Создаем тестовые PVZ
		for i := 0; i < 3; i++ {
			err := service.Create(ctx, &pvz.PVZ{
				Name:    fmt.Sprintf("Test PVZ %d", i),
				Address: fmt.Sprintf("Test Address %d", i),
			})
			require.NoError(t, err)
		}

		// Получаем список PVZ с дефолтной пагинацией
		list, err := service.GetPvzList(ctx, 0, 0, nil)
		require.NoError(t, err)
		assert.Len(t, list, 3)
	})

	t.Run("success with custom pagination", func(t *testing.T) {
		repo := repository.NewMockRepo()
		txManager := repository.NewMockTxManager()
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		// Создаем тестовые PVZ
		for i := 0; i < 5; i++ {
			err := service.Create(ctx, &pvz.PVZ{
				Name:    fmt.Sprintf("Test PVZ %d", i),
				Address: fmt.Sprintf("Test Address %d", i),
			})
			require.NoError(t, err)
		}

		// Получаем список PVZ с кастомной пагинацией
		list, err := service.GetPvzList(ctx, 1, 2, nil)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("success with filters", func(t *testing.T) {
		repo := repository.NewMockRepo()
		txManager := repository.NewMockTxManager()
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		// Создаем тестовые PVZ
		testPVZ := &pvz.PVZ{
			Name:    "Filtered PVZ",
			Address: "Filtered Address",
			Status:  pvz.StatusActive,
		}
		err := service.Create(ctx, testPVZ)
		require.NoError(t, err)

		// Создаем еще один PVZ с другими данными
		otherPVZ := &pvz.PVZ{
			Name:    "Other PVZ",
			Address: "Other Address",
			Status:  pvz.StatusInactive,
		}
		err = service.Create(ctx, otherPVZ)
		require.NoError(t, err)

		// Получаем список PVZ с фильтрами
		filters := map[string]interface{}{
			"name":    "Filtered PVZ",
			"address": "Filtered Address",
			"status":  pvz.StatusActive,
		}
		list, err := service.GetPvzList(ctx, 0, 10, filters)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, testPVZ.Name, list[0].Name)
	})

	t.Run("begin tx error", func(t *testing.T) {
		repo := repository.NewMockRepo()
		txManager := repository.NewMockTxManager()
		txManager.BeginTxError = errMock
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		_, err := service.GetPvzList(ctx, 0, 10, nil)
		assert.ErrorIs(t, err, errMock)
	})

	t.Run("list error with rollback", func(t *testing.T) {
		repo := repository.NewMockRepo()
		repo.ListError = errMock
		txManager := repository.NewMockTxManager()
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		_, err := service.GetPvzList(ctx, 0, 10, nil)
		assert.ErrorIs(t, err, errMock)
	})

	t.Run("commit error", func(t *testing.T) {
		repo := repository.NewMockRepo()
		txManager := repository.NewMockTxManager()
		txManager.CommitTxError = errMock
		service := NewPVZService(repo, txManager)
		ctx := context.Background()

		_, err := service.GetPvzList(ctx, 0, 10, nil)
		assert.ErrorIs(t, err, errMock)
	})
}

func TestMatchesFilters(t *testing.T) {
	testPVZ := &pvz.PVZ{
		Name:    "Test PVZ",
		Address: "Test Address",
		Status:  pvz.StatusActive,
	}

	t.Run("matches all filters", func(t *testing.T) {
		filters := map[string]interface{}{
			"name":    "Test PVZ",
			"address": "Test Address",
			"status":  pvz.StatusActive,
		}
		assert.True(t, matchesFilters(testPVZ, filters))
	})

	t.Run("does not match name filter", func(t *testing.T) {
		filters := map[string]interface{}{
			"name": "Different Name",
		}
		assert.False(t, matchesFilters(testPVZ, filters))
	})

	t.Run("does not match address filter", func(t *testing.T) {
		filters := map[string]interface{}{
			"address": "Different Address",
		}
		assert.False(t, matchesFilters(testPVZ, filters))
	})

	t.Run("does not match status filter", func(t *testing.T) {
		filters := map[string]interface{}{
			"status": pvz.StatusInactive,
		}
		assert.False(t, matchesFilters(testPVZ, filters))
	})

	t.Run("empty filters", func(t *testing.T) {
		assert.True(t, matchesFilters(testPVZ, nil))
		assert.True(t, matchesFilters(testPVZ, map[string]interface{}{}))
	})

	t.Run("invalid filter type", func(t *testing.T) {
		filters := map[string]interface{}{
			"name": 123, // неправильный тип для имени
		}
		assert.True(t, matchesFilters(testPVZ, filters))
	})
}
