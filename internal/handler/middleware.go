package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/avito/pvz/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func (h *Handler) authMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
		c.Abort()
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
		c.Abort()
		return
	}

	tokenString := parts[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.config.JWTConfig.Secret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		c.Abort()
		return
	}

	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
		c.Abort()
		return
	}

	role := models.UserRole(claims["role"].(string))
	log.Printf("Token role: %s", role)
	if role != models.RoleEmployee && role != models.RoleModerator {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user role"})
		c.Abort()
		return
	}

	c.Set("userID", userID)
	c.Set("userRole", string(role))
	log.Printf("Set user role in context: %s", string(role))
	c.Next()
}
