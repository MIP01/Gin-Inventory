package helper

import (
	"Gin-Inventory/middleware"

	"github.com/gin-gonic/gin"
)

// ValidationHelper adalah fungsi helper untuk memvalidasi input request
func ValidationHelper[T any](c *gin.Context, schema T) (T, bool) {
	// Bind JSON ke struct
	if err := c.ShouldBindJSON(&schema); err != nil {
		errors := middleware.FormatValidationErrors(err)
		c.JSON(400, gin.H{"errors": errors})
		return schema, false
	}

	// Validasi menggunakan ValidateInput
	if validationErrors := middleware.ValidateInput(schema); validationErrors != nil {
		c.JSON(400, gin.H{"errors": validationErrors})
		return schema, false
	}

	return schema, true
}
