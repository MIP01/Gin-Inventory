package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware mengatur header CORS untuk mengizinkan akses lintas domain
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Tambahkan header CORS
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // Ubah "*" ke domain tertentu jika perlu
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Jika request adalah OPTIONS, berhenti di sini
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
