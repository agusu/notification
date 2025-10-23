package middleware

import (
	"context"
	"fmt"
	"net/http"
	"notification/models"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserProvider interface {
	GetById(ctx context.Context, id uint) (models.User, error)
}

func AuthMiddleware(up UserProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString, found := strings.CutPrefix(authHeader, "Bearer ")
		if !found {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected method: %s", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if float64(claims["exp"].(float64)) < float64(time.Now().Unix()) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				c.Abort()
				return
			}
			user, err := up.GetById(c.Request.Context(), uint(claims["sub"].(float64)))
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			c.Set("user", user)
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
