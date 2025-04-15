package middleware

import (
	"net/http"
	"pec2-backend/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func extractJwtClaims(c *gin.Context) (jwt.MapClaims, bool) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		c.Abort()
		return nil, false
	}

	authHeader = strings.Trim(authHeader, "\"' ")

	if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		authHeader = "Bearer " + authHeader
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format, expected: Bearer <token>"})
		c.Abort()
		return nil, false
	}

	tokenString := parts[1]
	tokenString = strings.Trim(tokenString, "\"' ")

	claims, err := utils.DecodeJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token: " + err.Error()})
		c.Abort()
		return nil, false
	}

	return claims, true
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := extractJwtClaims(c)
		if !ok {
			return
		}

		c.Set("user_id", claims["user_id"])
		c.Set("role", claims["role"])
		c.Next()
	}
}

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := extractJwtClaims(c)
		if !ok {
			return
		}

		c.Set("user_id", claims["user_id"])
		c.Set("role", claims["role"])

		role, exists := claims["role"]
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Role not found in token"})
			c.Abort()
			return
		}

		if role != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: admin role required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
