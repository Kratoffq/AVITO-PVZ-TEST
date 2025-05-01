package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/avito/pvz/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
	}{
		{
			name: "successful creation",
			user: &models.User{
				ID:        uuid.New(),
				Email:     "test@example.com",
				Password:  "hashed_password",
				Role:      models.RoleEmployee,
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			user: &models.User{
				ID:        uuid.New(),
				Email:     "test@example.com", // Тот же email, что и в первом тесте
				Password:  "hashed_password",
				Role:      models.RoleEmployee,
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateUser(ctx, tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Проверяем, что пользователь создался
				created, err := repo.GetUserByEmail(ctx, tt.user.Email)
				require.NoError(t, err)
				assert.Equal(t, tt.user.ID, created.ID)
				assert.Equal(t, tt.user.Email, created.Email)
				assert.Equal(t, tt.user.Role, created.Role)
			}
		})
	}
}

func TestRepository_GetUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Создаем тестового пользователя
	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "hashed_password",
		Role:      models.RoleEmployee,
		CreatedAt: time.Now(),
	}
	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name    string
		email   string
		want    *models.User
		wantErr bool
	}{
		{
			name:  "existing user",
			email: user.Email,
			want:  user,
		},
		{
			name:  "non-existing user",
			email: "nonexistent@example.com",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetUserByEmail(ctx, tt.email)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.want == nil {
					assert.Nil(t, got)
				} else {
					assert.Equal(t, tt.want.ID, got.ID)
					assert.Equal(t, tt.want.Email, got.Email)
					assert.Equal(t, tt.want.Role, got.Role)
				}
			}
		})
	}
}

func TestRepository_GetUserByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Создаем тестового пользователя
	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "hashed_password",
		Role:      models.RoleEmployee,
		CreatedAt: time.Now(),
	}
	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		want    *models.User
		wantErr bool
	}{
		{
			name: "existing user",
			id:   user.ID,
			want: user,
		},
		{
			name: "non-existing user",
			id:   uuid.New(),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetUserByID(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.want == nil {
					assert.Nil(t, got)
				} else {
					assert.Equal(t, tt.want.ID, got.ID)
					assert.Equal(t, tt.want.Email, got.Email)
					assert.Equal(t, tt.want.Role, got.Role)
				}
			}
		})
	}
}
