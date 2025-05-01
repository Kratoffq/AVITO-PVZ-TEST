package postgres

import (
	"context"
	"database/sql"

	"github.com/avito/pvz/internal/models"
	"github.com/google/uuid"
)

// CreateUser создает нового пользователя
func (r *Repository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, password, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Password,
		user.Role,
		user.CreatedAt,
	)

	return err
}

// GetUserByEmail получает пользователя по email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password, role, created_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID получает пользователя по ID
func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, password, role, created_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
