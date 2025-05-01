package auth

import (
	"errors"
	"time"

	"github.com/avito/pvz/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	// SecretKey используется для подписи JWT токенов
	SecretKey = "your-secret-key"
)

var (
	// ErrInvalidToken возвращается при невалидном токене
	ErrInvalidToken = errors.New("invalid token")

	// ErrInvalidPassword возвращается при неверном пароле
	ErrInvalidPassword = errors.New("invalid password")
)

// Claims представляет собой данные, хранящиеся в JWT токене
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   user.Role `json:"role"`
	jwt.RegisteredClaims
}

// HashPassword хеширует пароль
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash проверяет соответствие пароля и хеша
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken генерирует JWT токен
func GenerateToken(userID uuid.UUID, role user.Role) (string, error) {
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(SecretKey))
}

// ValidateToken проверяет валидность JWT токена
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
