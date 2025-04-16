package repository

import (
	"database/sql"
)

type PVZRepository struct {
	db *sql.DB
}

func NewPVZRepository(db *sql.DB) *PVZRepository {
	return &PVZRepository{
		db: db,
	}
}

// TODO: Implement repository methods
