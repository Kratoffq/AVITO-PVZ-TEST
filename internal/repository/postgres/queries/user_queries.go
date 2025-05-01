package queries

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/avito/pvz/internal/domain/user"
	"github.com/google/uuid"
)

// CreateUser создает нового пользователя
func CreateUser(id uuid.UUID, email, password string, role user.Role, createdAt time.Time) (string, []interface{}, error) {
	return PostgresBuilder.Insert("users").
		Columns("id", "email", "password", "role", "created_at").
		Values(FormatUUID(id), email, password, role, createdAt).
		ToSql()
}

// GetUserByID получает пользователя по ID
func GetUserByID(id uuid.UUID) (string, []interface{}, error) {
	return PostgresBuilder.Select("id", "email", "password", "role", "created_at").
		From("users").
		Where(squirrel.Eq{"id": FormatUUID(id)}).
		ToSql()
}

// GetUserByEmail получает пользователя по email
func GetUserByEmail(email string) (string, []interface{}, error) {
	return PostgresBuilder.Select("id", "email", "password", "role", "created_at").
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()
}

// UpdateUser обновляет данные пользователя
func UpdateUser(id uuid.UUID, email, password string, role user.Role) (string, []interface{}, error) {
	return PostgresBuilder.Update("users").
		Set("email", email).
		Set("password", password).
		Set("role", role).
		Where(squirrel.Eq{"id": FormatUUID(id)}).
		ToSql()
}

// DeleteUser удаляет пользователя
func DeleteUser(id uuid.UUID) (string, []interface{}, error) {
	return PostgresBuilder.Delete("users").
		Where(squirrel.Eq{"id": FormatUUID(id)}).
		ToSql()
}

// ListUsers получает список пользователей с пагинацией
func ListUsers(offset, limit int) (string, []interface{}, error) {
	return Paginate(
		PostgresBuilder.Select("id", "email", "password", "role", "created_at").
			From("users").
			OrderBy("created_at DESC"),
		offset,
		limit,
	)
}
