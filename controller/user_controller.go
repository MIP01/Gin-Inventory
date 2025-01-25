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

func CreateUserHandler(c *gin.Context) {
	// Memvalidasi input dengan Middleware ValidateInput.
	userData, valid := helper.ValidationHelper(c, middleware.UserSchema{})
	if !valid {
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
	id := c.Param("id")

	// handle role
	currentUserID, role, valid := helper.CheckUserRoleAndID(c, "user", "admin")
	if !valid {
		return
	}

	// Pastikan user ada
	var user model.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Periksa akses berdasarkan role
	if role == "user" {
		// Pastikan user hanya bisa mengakses datanya sendiri
		if fmt.Sprintf("%v", currentUserID) != id {
			c.JSON(403, gin.H{"error": "Forbidden: You can only access your own data"})
			return
		}
	}

	c.JSON(200, user.ToMap())
}

func UpdateUserHandler(c *gin.Context) {
	id := c.Param("id")

	// handle role
	currentUserID, role, valid := helper.CheckUserRoleAndID(c, "user", "admin")
	if !valid {
		return
	}

	// Pastikan user ada
	var user model.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Periksa akses berdasarkan role
	if role == "user" {
		// Pastikan user hanya bisa mengakses datanya sendiri
		if fmt.Sprintf("%v", currentUserID) != id {
			c.JSON(403, gin.H{"error": "Forbidden: You can only access your own data"})
			return
		}
	}

	// Memvalidasi input dengan Middleware ValidateInput.
	updatedData, valid := helper.ValidationHelper(c, middleware.UpdateSchema{})
	if !valid {
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

	// handle role
	_, _, valid := helper.CheckUserRoleAndID(c, "admin")
	if !valid {
		return
	}

	// Pastikan user ada
	var user model.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Mulai transaksi database
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Failed to delete user and related data"})
		}
	}()

	// Cari semua transaksi milik user
	var transactions []model.Transaction
	if err := tx.Where("user_id = ?", id).Find(&transactions).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Failed to find user transactions"})
		return
	}

	// Proses jika status detail adalah "loaned", kembalikan stok item
	for _, transaction := range transactions {
		if transaction.DetailID == nil {
			continue
		}

		var detail model.Detail
		if err := tx.First(&detail, *transaction.DetailID).Error; err != nil {
			tx.Rollback()
			c.JSON(404, gin.H{"error": "Detail not found"})
			return
		}

		if detail.Status == "loaned" {
			var item model.Item
			if err := tx.First(&item, transaction.ItemID).Error; err != nil {
				tx.Rollback()
				c.JSON(404, gin.H{"error": "Item not found"})
				return
			}

			item.Stock += transaction.Quantity
			if err := tx.Save(&item).Error; err != nil {
				tx.Rollback()
				c.JSON(500, gin.H{"error": "Failed to update item stock"})
				return
			}
		}
	}

	// Hapus detail terkait
	var detailIDs []uint
	for _, transaction := range transactions {
		if transaction.DetailID != nil {
			detailIDs = append(detailIDs, *transaction.DetailID)
		}
	}
	if len(detailIDs) > 0 {
		if err := tx.Unscoped().Where("id IN (?)", detailIDs).Delete(&model.Detail{}).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Failed to delete user details"})
			return
		}
	}

	// Hapus transaksi terkait
	if err := tx.Unscoped().Where("user_id = ?", id).Delete(&model.Transaction{}).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Failed to delete user transactions"})
		return
	}

	// Hapus user
	if err := tx.Unscoped().Delete(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Failed to delete user"})
		return
	}

	tx.Commit()

	c.JSON(200, gin.H{"message": "User and related data deleted successfully"})
}
