package controller

import (
	"fmt"
	"gin_iventory/config"
	"gin_iventory/middleware"
	"gin_iventory/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserHandler(c *gin.Context) {
	var userData middleware.UserSchema
	if err := c.ShouldBindJSON(&userData); err != nil {
		errors := middleware.FormatValidationErrors(err)
		c.JSON(400, gin.H{"errors": errors})
		return
	}

	validationErrors := middleware.ValidateInput(userData)
	if validationErrors != nil {
		c.JSON(400, gin.H{"errors": validationErrors})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}

	userData.Password = string(hashedPassword)

	newUser := model.User{
		Name:     userData.Name,
		Email:    userData.Email,
		Password: userData.Password,
		Role:     "user", // Atur default role
	}

	if err := config.DB.Create(&newUser).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "User created successfully", "user": newUser.ToMap()})
}

func GetAllUserHandler(c *gin.Context) {
	var user []model.User
	if err := config.DB.Find(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"user": model.UsersToMap(user)})
}

func GetUserHandler(c *gin.Context) {
	// Ambil ID user dari parameter URL
	id := c.Param("id")

	// Ambil current_id dan role dari context (diset oleh middleware)
	currentUserID, exists := c.Get("current_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	role, roleExists := c.Get("role")
	if !roleExists || (role != "admin" && id != fmt.Sprint(currentUserID)) {
		c.JSON(403, gin.H{"error": "Forbidden: You can only access your own data"})
		return
	}

	// Pastikan user ada
	var user model.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	c.JSON(200, user.ToMap())
}

func UpdateUserHandler(c *gin.Context) {
	id := c.Param("id")

	// Ambil current_id dan role dari context (diset oleh middleware)
	currentUserID, exists := c.Get("current_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	role, roleExists := c.Get("role")
	if !roleExists || (role != "admin" && id != fmt.Sprint(currentUserID)) {
		c.JSON(403, gin.H{"error": "Forbidden: You can only access your own data"})
		return
	}

	// Pastikan user ada
	var user model.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Memvalidasi input dengan Middleware ValidateInput.
	var updatedData middleware.UpdateSchema
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		errors := middleware.FormatValidationErrors(err)
		c.JSON(400, gin.H{"errors": errors})
		return
	}

	validationErrors := middleware.ValidateInput(updatedData)
	if validationErrors != nil {
		c.JSON(400, gin.H{"errors": validationErrors})
		return
	}

	// Perbarui field yang diberikan
	if updatedData.Name != "" {
		user.Name = updatedData.Name
	}
	if updatedData.Email != "" {
		user.Email = updatedData.Email
	}
	// Jika password diperbarui, hash terlebih dahulu
	if updatedData.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatedData.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = string(hashedPassword)
	}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "User updated successfully", "user": user.ToMap()})
}

func DeleteUserHandler(c *gin.Context) {
	id := c.Param("id")

	// Ambil current_id dan role dari context (diset oleh middleware)
	currentUserID, exists := c.Get("current_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	role, roleExists := c.Get("role")
	if !roleExists || (role != "admin" && id != fmt.Sprint(currentUserID)) {
		c.JSON(403, gin.H{"error": "Forbidden: You can only access your own data"})
		return
	}

	// Pastikan user ada
	var user model.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Hapus user
	if err := config.DB.Delete(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "User deleted successfully"})
}
