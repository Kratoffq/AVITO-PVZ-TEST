package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/avito/pvz/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullPVZFlow(t *testing.T) {
	// Инициализация тестового окружения
	ctx := context.Background()
	db, err := setupTestDB()
	require.NoError(t, err)
	defer db.Close()

	t.Run("Полный интеграционный тест ПВЗ: создание ПВЗ, приёмки и обработка товаров", func(t *testing.T) {
		// Начинаем транзакцию
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)
		defer tx.Rollback()

		// 1. Создаем новый ПВЗ
		pvzID := uuid.New()
		pvzCity := "Москва"
		query := `INSERT INTO pvzs (id, registration_date, city) VALUES ($1, $2, $3)`
		_, err = tx.ExecContext(ctx, query, pvzID, time.Now(), pvzCity)
		require.NoError(t, err)

		// Проверяем создание ПВЗ
		var city string
		query = `SELECT city FROM pvzs WHERE id = $1`
		err = tx.QueryRowContext(ctx, query, pvzID).Scan(&city)
		require.NoError(t, err)
		assert.Equal(t, pvzCity, city)

		// 2. Создаем новую приёмку
		receptionID := uuid.New()
		query = `INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, $2, $3, $4)`
		_, err = tx.ExecContext(ctx, query, receptionID, time.Now(), pvzID, models.StatusInProgress)
		require.NoError(t, err)

		// 3. Добавляем 50 товаров
		for i := 0; i < 50; i++ {
			query = `INSERT INTO products (id, date_time, type, reception_id) VALUES ($1, $2, $3, $4)`
			_, err = tx.ExecContext(ctx, query, uuid.New(), time.Now(), models.TypeElectronics, receptionID)
			require.NoError(t, err)
		}

		// Проверяем общее количество товаров
		var totalCount int
		query = `SELECT COUNT(*) FROM products WHERE reception_id = $1`
		err = tx.QueryRowContext(ctx, query, receptionID).Scan(&totalCount)
		require.NoError(t, err)
		assert.Equal(t, 50, totalCount)

		// 4. Закрываем приёмку
		query = `UPDATE receptions SET status = $1 WHERE id = $2`
		_, err = tx.ExecContext(ctx, query, models.StatusClose, receptionID)
		require.NoError(t, err)

		// Проверяем статус приёмки
		var status models.ReceptionStatus
		query = `SELECT status FROM receptions WHERE id = $1`
		err = tx.QueryRowContext(ctx, query, receptionID).Scan(&status)
		require.NoError(t, err)
		assert.Equal(t, models.StatusClose, status)

		// Фиксируем транзакцию
		err = tx.Commit()
		require.NoError(t, err)
	})
}

func TestPVZPagination(t *testing.T) {
	ctx := context.Background()
	db, err := setupTestDB()
	require.NoError(t, err)
	defer db.Close()

	t.Run("Тест пагинации ПВЗ", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)
		defer tx.Rollback()

		// Создаем 15 ПВЗ
		for i := 0; i < 15; i++ {
			query := `INSERT INTO pvzs (id, registration_date, city) VALUES ($1, $2, $3)`
			_, err = tx.ExecContext(ctx, query, uuid.New(), time.Now(), fmt.Sprintf("Город %d", i))
			require.NoError(t, err)
		}

		// Проверяем первую страницу (10 записей)
		var count int
		query := `SELECT COUNT(*) FROM (SELECT * FROM pvzs ORDER BY registration_date DESC LIMIT 10 OFFSET 0) as p`
		err = tx.QueryRowContext(ctx, query).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 10, count)

		// Проверяем вторую страницу (5 записей)
		query = `SELECT COUNT(*) FROM (SELECT * FROM pvzs ORDER BY registration_date DESC LIMIT 10 OFFSET 10) as p`
		err = tx.QueryRowContext(ctx, query).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 5, count)

		err = tx.Commit()
		require.NoError(t, err)
	})
}

func TestPVZDateFiltering(t *testing.T) {
	ctx := context.Background()
	db, err := setupTestDB()
	require.NoError(t, err)
	defer db.Close()

	t.Run("Тест фильтрации ПВЗ по датам", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)
		defer tx.Rollback()

		// Создаем ПВЗ с разными датами
		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)
		tomorrow := now.Add(24 * time.Hour)

		pvz1 := uuid.New()
		pvz2 := uuid.New()
		pvz3 := uuid.New()

		query := `INSERT INTO pvzs (id, registration_date, city) VALUES ($1, $2, $3)`
		_, err = tx.ExecContext(ctx, query, pvz1, yesterday, "Вчерашний город")
		require.NoError(t, err)

		_, err = tx.ExecContext(ctx, query, pvz2, now, "Сегодняшний город")
		require.NoError(t, err)

		_, err = tx.ExecContext(ctx, query, pvz3, tomorrow, "Завтрашний город")
		require.NoError(t, err)

		// Проверяем фильтрацию по диапазону дат
		var count int
		query = `SELECT COUNT(*) FROM pvzs WHERE registration_date BETWEEN $1 AND $2`
		err = tx.QueryRowContext(ctx, query, yesterday, now).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count)

		err = tx.Commit()
		require.NoError(t, err)
	})
}

func TestPVZErrorHandling(t *testing.T) {
	ctx := context.Background()
	db, err := setupTestDB()
	require.NoError(t, err)
	defer db.Close()

	t.Run("Тест обработки ошибок ПВЗ", func(t *testing.T) {
		// Тест на дублирование ID
		pvzID := uuid.New()
		query := `INSERT INTO pvzs (id, registration_date, city) VALUES ($1, $2, $3)`
		_, err = db.ExecContext(ctx, query, pvzID, time.Now(), "Город")
		require.NoError(t, err)

		// Пытаемся создать ПВЗ с тем же ID
		_, err = db.ExecContext(ctx, query, pvzID, time.Now(), "Другой город")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate key")

		// Тест на неверный формат UUID
		query = `SELECT * FROM pvzs WHERE id = $1`
		_, err = db.QueryContext(ctx, query, "неверный-uuid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid input syntax")

		// Тест на несуществующий ПВЗ
		query = `SELECT * FROM pvzs WHERE id = $1`
		var city string
		err = db.QueryRowContext(ctx, query, uuid.New()).Scan(&city)
		assert.Error(t, err)
		assert.Equal(t, "sql: no rows in result set", err.Error())
	})
}
