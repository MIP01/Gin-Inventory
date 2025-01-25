package controller

import (
	"Gin-Inventory/config"
	"Gin-Inventory/helper"
	"Gin-Inventory/middleware"
	"Gin-Inventory/model"
	"fmt"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func CreateAdminHandler(c *gin.Context) {
	// Memvalidasi input dengan Middleware ValidateInput.
	adminData, valid := helper.ValidationHelper(c, middleware.UserSchema{})
	if !valid {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminData.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}

	adminData.Password = string(hashedPassword)

	newAdmin := model.Admin{
		Name:     adminData.Name,
		Email:    adminData.Email,
		Password: adminData.Password,
		Role:     "admin", // Atur default role
	}

	if err := config.DB.Create(&newAdmin).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Admin created successfully", "admin": newAdmin.ToMap()})
}

func GetAllAdminHandler(c *gin.Context) {
	var admin []model.Admin
	if err := config.DB.Find(&admin).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"admin": model.AdminsToMap(admin)})
}

func GetAdminHandler(c *gin.Context) {
	id := c.Param("id")

	// handle role
	currentUserID, _, valid := helper.CheckUserRoleAndID(c, "admin")
	if !valid {
		return
	}

	// Pastikan admin ada
	var admin model.Admin
	if err := config.DB.First(&admin, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Admin not found"})
		return
	}

	// Pastikan admin hanya bisa mengakses datanya sendiri
	if fmt.Sprintf("%v", currentUserID) != id {
		c.JSON(403, gin.H{"error": "Forbidden: You can only access your own data"})
		return
	}

	c.JSON(200, admin.ToMap())
}

func UpdateAdminHandler(c *gin.Context) {
	id := c.Param("id")

	// handle role
	currentUserID, _, valid := helper.CheckUserRoleAndID(c, "admin")
	if !valid {
		return
	}

	// Pastikan admin ada
	var admin model.Admin
	if err := config.DB.First(&admin, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Admin not found"})
		return
	}

	// Pastikan admin hanya bisa mengakses datanya sendiri
	if fmt.Sprintf("%v", currentUserID) != id {
		c.JSON(403, gin.H{"error": "Forbidden: You can only access your own data"})
		return
	}

	// Memvalidasi input dengan Middleware ValidateInput.
	updatedData, valid := helper.ValidationHelper(c, middleware.UpdateSchema{})
	if !valid {
		return
	}

	// Perbarui field yang diberikan
	if updatedData.Name != "" {
		admin.Name = updatedData.Name
	}
	if updatedData.Email != "" {
		admin.Email = updatedData.Email
	}
	// Jika password diperbarui, hash terlebih dahulu
	if updatedData.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatedData.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to hash password"})
			return
		}
		admin.Password = string(hashedPassword)
	}

	if err := config.DB.Save(&admin).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Admin updated successfully", "admin": admin.ToMap()})
}

func DeleteAdminHandler(c *gin.Context) {
	id := c.Param("id")

	// handle role
	currentUserID, _, valid := helper.CheckUserRoleAndID(c, "admin")
	if !valid {
		return
	}

	// Pastikan admin ada
	var admin model.Admin
	if err := config.DB.First(&admin, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Admin not found"})
		return
	}

	// Pastikan admin hanya bisa mengakses datanya sendiri
	if fmt.Sprintf("%v", currentUserID) != id {
		c.JSON(403, gin.H{"error": "Forbidden: You can only access your own data"})
		return
	}

	if err := config.DB.Unscoped().Delete(&admin).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Admin deleted successfully"})
}
