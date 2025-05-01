package repository

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBaseRepository(t *testing.T) {
	db := &sql.DB{}
	repo := NewBaseRepository(db)

	assert.Equal(t, db, repo.db)
	assert.Nil(t, repo.tx)
}

func TestBaseRepository_GetDB(t *testing.T) {
	db := &sql.DB{}
	tx := &sql.Tx{}

	tests := []struct {
		name string
		repo BaseRepository
		want interface{}
	}{
		{
			name: "возвращает БД когда нет транзакции",
			repo: BaseRepository{db: db},
			want: db,
		},
		{
			name: "возвращает транзакцию когда она есть",
			repo: BaseRepository{db: db, tx: tx},
			want: tx,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.repo.GetDB()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBaseRepository_WithTx(t *testing.T) {
	db := &sql.DB{}
	tx := &sql.Tx{}

	repo := NewBaseRepository(db)
	repoWithTx := repo.WithTx(tx)

	// Проверяем, что исходный репозиторий не изменился
	assert.Equal(t, db, repo.db)
	assert.Nil(t, repo.tx)

	// Проверяем, что новый репозиторий имеет транзакцию
	assert.Equal(t, db, repoWithTx.db)
	assert.Equal(t, tx, repoWithTx.tx)
}
