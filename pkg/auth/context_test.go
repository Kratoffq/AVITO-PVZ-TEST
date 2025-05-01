package auth

import (
	"context"
	"testing"

	"github.com/avito/pvz/internal/domain/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWithUserID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	// Проверяем добавление ID пользователя
	newCtx := WithUserID(ctx, userID)
	assert.NotEqual(t, ctx, newCtx)

	// Проверяем получение ID пользователя
	retrievedID, ok := GetUserID(newCtx)
	assert.True(t, ok)
	assert.Equal(t, userID, retrievedID)

	// Проверяем отсутствие ID в исходном контексте
	_, ok = GetUserID(ctx)
	assert.False(t, ok)
}

func TestWithUserRole(t *testing.T) {
	ctx := context.Background()
	role := user.Role("admin")

	// Проверяем добавление роли пользователя
	newCtx := WithUserRole(ctx, role)
	assert.NotEqual(t, ctx, newCtx)

	// Проверяем получение роли пользователя
	retrievedRole, ok := GetUserRole(newCtx)
	assert.True(t, ok)
	assert.Equal(t, role, retrievedRole)

	// Проверяем отсутствие роли в исходном контексте
	_, ok = GetUserRole(ctx)
	assert.False(t, ok)
}

func TestContextWithBothValues(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	role := user.Role("admin")

	// Добавляем оба значения
	ctx = WithUserID(ctx, userID)
	ctx = WithUserRole(ctx, role)

	// Проверяем оба значения
	retrievedID, ok := GetUserID(ctx)
	assert.True(t, ok)
	assert.Equal(t, userID, retrievedID)

	retrievedRole, ok := GetUserRole(ctx)
	assert.True(t, ok)
	assert.Equal(t, role, retrievedRole)
}
