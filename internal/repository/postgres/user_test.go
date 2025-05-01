package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/avito/pvz/internal/domain/user"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	u := &user.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "password123",
		Role:      user.RoleAdmin,
		CreatedAt: time.Now(),
	}

	mock.ExpectExec("INSERT INTO users").
		WithArgs(u.ID, u.Email, u.Password, u.Role, u.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(context.Background(), u)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	tests := []struct {
		name        string
		setup       func() (uuid.UUID, *user.User)
		expectedErr error
	}{
		{
			name: "успешное получение пользователя",
			setup: func() (uuid.UUID, *user.User) {
				userID := uuid.New()
				createdAt := time.Now()
				u := &user.User{
					ID:        userID,
					Email:     "test@example.com",
					Password:  "password123",
					Role:      user.RoleAdmin,
					CreatedAt: createdAt,
				}

				rows := sqlmock.NewRows([]string{"id", "email", "password", "role", "created_at"}).
					AddRow(userID, u.Email, u.Password, u.Role, createdAt)
				mock.ExpectQuery("SELECT id, email, password, role, created_at FROM users").
					WithArgs(userID).
					WillReturnRows(rows)

				return userID, u
			},
			expectedErr: nil,
		},
		{
			name: "пользователь не найден",
			setup: func() (uuid.UUID, *user.User) {
				userID := uuid.New()
				mock.ExpectQuery("SELECT id, email, password, role, created_at FROM users").
					WithArgs(userID).
					WillReturnError(sql.ErrNoRows)
				return userID, nil
			},
			expectedErr: user.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, expected := tt.setup()

			result, err := repo.GetByID(context.Background(), id)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expected.Email, result.Email)
				assert.Equal(t, expected.Password, result.Password)
				assert.Equal(t, expected.Role, result.Role)
				assert.Equal(t, expected.ID, result.ID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	tests := []struct {
		name        string
		email       string
		mockSetup   func()
		expected    *user.User
		expectedErr error
	}{
		{
			name:  "успешное получение пользователя",
			email: "test@example.com",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "email", "password", "role", "created_at"}).
					AddRow(uuid.New(), "test@example.com", "password123", user.RoleAdmin, time.Now())
				mock.ExpectQuery("SELECT id, email, password, role, created_at FROM users").
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expected: &user.User{
				ID:        uuid.New(),
				Email:     "test@example.com",
				Password:  "password123",
				Role:      user.RoleAdmin,
				CreatedAt: time.Now(),
			},
			expectedErr: nil,
		},
		{
			name:  "пользователь не найден",
			email: "nonexistent@example.com",
			mockSetup: func() {
				mock.ExpectQuery("SELECT id, email, password, role, created_at FROM users").
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			expected:    nil,
			expectedErr: user.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.GetByEmail(context.Background(), tt.email)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Email, result.Email)
				assert.Equal(t, tt.expected.Password, result.Password)
				assert.Equal(t, tt.expected.Role, result.Role)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	u := &user.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "password123",
		Role:      user.RoleAdmin,
		CreatedAt: time.Now(),
	}

	mock.ExpectExec("UPDATE users SET").
		WithArgs(u.Email, u.Password, u.Role, u.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(context.Background(), u)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	userID := uuid.New()

	mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(context.Background(), userID)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	tests := []struct {
		name        string
		offset      int
		limit       int
		mockSetup   func() []*user.User
		expectedErr error
	}{
		{
			name:   "успешное получение списка пользователей",
			offset: 0,
			limit:  10,
			mockSetup: func() []*user.User {
				users := []*user.User{
					{
						ID:        uuid.MustParse("a01674f4-1331-49db-a445-0fe37c78dd68"),
						Email:     "test1@example.com",
						Password:  "password123",
						Role:      user.RoleAdmin,
						CreatedAt: time.Date(2025, 5, 1, 22, 24, 21, 957283626, time.Local),
					},
					{
						ID:        uuid.MustParse("c21549f4-3a8f-47e2-a7a3-774bf88a09d4"),
						Email:     "test2@example.com",
						Password:  "password456",
						Role:      user.RoleEmployee,
						CreatedAt: time.Date(2025, 5, 1, 22, 24, 21, 957284376, time.Local),
					},
				}

				rows := sqlmock.NewRows([]string{"id", "email", "password", "role", "created_at"})
				for _, u := range users {
					rows.AddRow(u.ID, u.Email, u.Password, u.Role, u.CreatedAt)
				}

				mock.ExpectQuery("SELECT id, email, password, role, created_at FROM users ORDER BY created_at DESC LIMIT 10 OFFSET 0").
					WillReturnRows(rows)

				return users
			},
			expectedErr: nil,
		},
		{
			name:   "пустой список",
			offset: 0,
			limit:  10,
			mockSetup: func() []*user.User {
				rows := sqlmock.NewRows([]string{"id", "email", "password", "role", "created_at"})
				mock.ExpectQuery("SELECT id, email, password, role, created_at FROM users ORDER BY created_at DESC LIMIT 10 OFFSET 0").
					WillReturnRows(rows)
				return []*user.User{}
			},
			expectedErr: nil,
		},
		{
			name:   "ошибка базы данных",
			offset: 0,
			limit:  10,
			mockSetup: func() []*user.User {
				mock.ExpectQuery("SELECT id, email, password, role, created_at FROM users ORDER BY created_at DESC LIMIT 10 OFFSET 0").
					WillReturnError(sql.ErrConnDone)
				return nil
			},
			expectedErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := tt.mockSetup()

			result, err := repo.List(context.Background(), tt.offset, tt.limit)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(expected), len(result))
				for i, u := range expected {
					assert.Equal(t, u.ID, result[i].ID)
					assert.Equal(t, u.Email, result[i].Email)
					assert.Equal(t, u.Password, result[i].Password)
					assert.Equal(t, u.Role, result[i].Role)
					assert.Equal(t, u.CreatedAt.Unix(), result[i].CreatedAt.Unix())
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	u := &user.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "password123",
		Role:      user.RoleAdmin,
		CreatedAt: time.Now(),
	}

	mock.ExpectExec("UPDATE users SET").
		WithArgs(u.Email, u.Password, u.Role, u.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(context.Background(), u)
	require.Error(t, err)
	require.Equal(t, user.ErrNotFound, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	userID := uuid.New()

	mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(context.Background(), userID)
	require.Error(t, err)
	require.Equal(t, user.ErrNotFound, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
