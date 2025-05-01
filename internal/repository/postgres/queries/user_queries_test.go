package queries

import (
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/avito/pvz/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "hashed_password",
		Role:      models.RoleEmployee,
		CreatedAt: time.Now(),
	}

	query, args, err := PostgresBuilder.
		Insert("users").
		Columns("id", "email", "password", "role", "created_at").
		Values(user.ID, user.Email, user.Password, user.Role, user.CreatedAt).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "INSERT INTO users (id,email,password,role,created_at) VALUES ($1,$2,$3,$4,$5)", query)
	assert.Len(t, args, 5)
	assert.Equal(t, user.ID, args[0])
	assert.Equal(t, user.Email, args[1])
	assert.Equal(t, user.Password, args[2])
	assert.Equal(t, user.Role, args[3])
	assert.Equal(t, user.CreatedAt, args[4])
}

func TestGetUserByID(t *testing.T) {
	id := uuid.New()
	query, args, err := PostgresBuilder.
		Select("id", "email", "password", "role", "created_at").
		From("users").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, email, password, role, created_at FROM users WHERE id = $1", query)
	assert.Equal(t, []interface{}{id.String()}, args)
}

func TestGetUserByEmail(t *testing.T) {
	email := "test@example.com"
	query, args, err := PostgresBuilder.
		Select("id", "email", "password", "role", "created_at").
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, email, password, role, created_at FROM users WHERE email = $1", query)
	assert.Equal(t, []interface{}{email}, args)
}

func TestUpdateUser(t *testing.T) {
	id := uuid.New()
	query, args, err := PostgresBuilder.
		Update("users").
		Set("email", "new@example.com").
		Set("password", "new_password").
		Set("role", models.RoleModerator).
		Where(squirrel.Eq{"id": id}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "UPDATE users SET email = $1, password = $2, role = $3 WHERE id = $4", query)
	assert.Equal(t, []interface{}{"new@example.com", "new_password", models.RoleModerator, id.String()}, args)
}

func TestDeleteUser(t *testing.T) {
	id := uuid.New()
	query, args, err := PostgresBuilder.
		Delete("users").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "DELETE FROM users WHERE id = $1", query)
	assert.Equal(t, []interface{}{id.String()}, args)
}

func TestListUsers(t *testing.T) {
	query, args, err := PostgresBuilder.
		Select("id", "email", "password", "role", "created_at").
		From("users").
		OrderBy("created_at DESC").
		Limit(20).
		Offset(10).
		ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, email, password, role, created_at FROM users ORDER BY created_at DESC LIMIT 20 OFFSET 10", query)
	assert.Empty(t, args)
}
