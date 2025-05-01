package postgres

import (
	"context"
	"database/sql"

	"github.com/avito/pvz/internal/domain/user"
	"github.com/avito/pvz/internal/repository/postgres/queries"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// UserRepository реализует интерфейс user.Repository
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create создает нового пользователя
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	query, args, err := queries.CreateUser(u.ID, u.Email, u.Password, u.Role, u.CreatedAt)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// GetByID получает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	query, args, err := queries.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	var result user.User
	err = r.db.GetContext(ctx, &result, query, args...)
	if err == sql.ErrNoRows {
		return nil, user.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetByEmail получает пользователя по email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query, args, err := queries.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	var result user.User
	err = r.db.GetContext(ctx, &result, query, args...)
	if err == sql.ErrNoRows {
		return nil, user.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Update обновляет данные пользователя
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	query, args, err := queries.UpdateUser(u.ID, u.Email, u.Password, u.Role)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return user.ErrNotFound
	}

	return nil
}

// Delete удаляет пользователя
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := queries.DeleteUser(id)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return user.ErrNotFound
	}

	return nil
}

// List возвращает список пользователей с пагинацией
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*user.User, error) {
	query, args, err := queries.ListUsers(offset, limit)
	if err != nil {
		return nil, err
	}

	var result []*user.User
	err = r.db.SelectContext(ctx, &result, query, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}
