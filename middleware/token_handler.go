package middleware

import (
	"fmt"
	"strings"

	"gin_iventory/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil token dari header Authorization
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(401, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Hilangkan prefix "Bearer " jika ada
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// Parse token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validasi metode signing
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(config.JWTSecret), nil
		})

		// Jika token valid, ambil klaim
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Ambil `user_id` dari klaim token
			currentID, ok := claims["user_id"].(float64) // `float64` karena nilai dari MapClaims default-nya float
			if !ok {
				c.JSON(401, gin.H{"error": "Invalid token claims"})
				c.Abort()
				return
			}

			// Simpan `current_id` ke dalam context
			c.Set("current_id", uint(currentID)) // Konversi ke `uint` sesuai tipe yang digunakan
			c.Set("role", claims["role"])        // Simpan juga role jika diperlukan
		} else {
			c.JSON(401, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Lanjutkan ke handler berikutnya
		c.Next()
	}
}
