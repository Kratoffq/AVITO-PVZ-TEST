package repository

import (
	"database/sql"
)

// TransactionalRepository определяет базовый интерфейс для репозиториев с поддержкой транзакций
type TransactionalRepository interface {
	// WithTx возвращает репозиторий, работающий в рамках переданной транзакции
	WithTx(tx *sql.Tx) TransactionalRepository
}

// BaseRepository предоставляет базовую реализацию для репозиториев
type BaseRepository struct {
	db *sql.DB
	tx *sql.Tx
}

// NewBaseRepository создает новый базовый репозиторий
func NewBaseRepository(db *sql.DB) BaseRepository {
	return BaseRepository{db: db}
}

// GetDB возвращает текущее соединение с БД или активную транзакцию
func (r *BaseRepository) GetDB() interface{} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// WithTx возвращает репозиторий с установленной транзакцией
func (r BaseRepository) WithTx(tx *sql.Tx) BaseRepository {
	r.tx = tx
	return r
}
