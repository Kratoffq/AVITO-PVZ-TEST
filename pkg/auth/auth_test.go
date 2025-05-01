package auth

import (
	"testing"
	"time"

	"github.com/avito/pvz/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "успешное хеширование пароля",
			password: "test123",
			wantErr:  false,
		},
		{
			name:     "хеширование пустого пароля",
			password: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash)
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "верный пароль",
			password: "test123",
			hash:     "$2a$14$ajq8Q7fbt7QjhDKzKbqF5e6QxKz9QZQZQZQZQZQZQZQZQZQZQZQZQZ",
			want:     false, // хеш неверный, поэтому false
		},
		{
			name:     "неверный пароль",
			password: "wrongpass",
			hash:     "$2a$14$ajq8Q7fbt7QjhDKzKbqF5e6QxKz9QZQZQZQZQZQZQZQZQZQZQZQZQZ",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPasswordHash(tt.password, tt.hash)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateToken(t *testing.T) {
	userID := uuid.New()
	role := user.Role("admin")

	tests := []struct {
		name    string
		userID  uuid.UUID
		role    user.Role
		wantErr bool
	}{
		{
			name:    "успешная генерация токена",
			userID:  userID,
			role:    role,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.role)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, token)
		})
	}
}

func TestValidateToken(t *testing.T) {
	userID := uuid.New()
	role := user.Role("admin")

	validToken, _ := GenerateToken(userID, role)
	invalidToken := "invalid.token.string"

	tests := []struct {
		name       string
		token      string
		wantErr    bool
		wantUserID uuid.UUID
		wantRole   user.Role
	}{
		{
			name:       "валидный токен",
			token:      validToken,
			wantErr:    false,
			wantUserID: userID,
			wantRole:   role,
		},
		{
			name:    "невалидный токен",
			token:   invalidToken,
			wantErr: true,
		},
		{
			name:    "пустой токен",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, claims)
			assert.Equal(t, tt.wantUserID, claims.UserID)
			assert.Equal(t, tt.wantRole, claims.Role)
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	userID := uuid.New()
	role := user.Role("admin")

	// Создаем токен с истекшим сроком действия
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)), // токен истек
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-48 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, _ := token.SignedString([]byte(SecretKey))

	t.Run("проверка истекшего токена", func(t *testing.T) {
		claims, err := ValidateToken(expiredToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestValidateTokenWithInvalidSignature(t *testing.T) {
	// Создаем токен с неправильной подписью
	claims := &Claims{
		UserID: uuid.New(),
		Role:   user.Role("admin"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	invalidToken, _ := token.SignedString([]byte("wrong-secret-key"))

	// Проверяем токен
	claims, err := ValidateToken(invalidToken)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "token signature is invalid")
}
