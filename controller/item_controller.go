package controller

import (
	"Gin-Inventory/config"
	"Gin-Inventory/helper"
	"Gin-Inventory/middleware"
	"Gin-Inventory/model"

	"github.com/gin-gonic/gin"
)

func CreateItemHandler(c *gin.Context) {
	// handle role
	_, _, valid := helper.CheckUserRoleAndID(c, "admin")
	if !valid {
		return
	}

	// Memvalidasi input dengan Middleware ValidateInput.
	itemData, valid := helper.ValidationHelper(c, middleware.ItemSchema{})
	if !valid {
		return
	}

	newItem := model.Item{
		Name:  itemData.Name,
		Stock: itemData.Stock,
	}

	if err := config.DB.Create(&newItem).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Item created successfully", "item": newItem.ToMap()})
}

func GetAllItemHandler(c *gin.Context) {
	var item []model.Item
	if err := config.DB.Find(&item).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"item": model.ItemsToMap(item)})
}

func GetItemHandler(c *gin.Context) {
	item_id := c.Param("item_id")

	// Pastikan item ada
	var item model.Item
	if err := config.DB.First(&item, item_id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(200, item.ToMap())
}

func UpdateItemHandler(c *gin.Context) {
	item_id := c.Param("item_id")

	// handle role
	_, _, valid := helper.CheckUserRoleAndID(c, "admin")
	if !valid {
		return
	}

	// Pastikan item ada
	var item model.Item
	if err := config.DB.First(&item, item_id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Item not found"})
		return
	}

	// Memvalidasi input dengan Middleware ValidateInput.
	updatedData, valid := helper.ValidationHelper(c, middleware.ItemSchema{})
	if !valid {
		return
	}

	// Perbarui field yang diberikan
	if updatedData.Name != "" {
		item.Name = updatedData.Name
	}
	if updatedData.Stock != 0 {
		item.Stock = updatedData.Stock
	}
	if err := config.DB.Save(&item).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Item updated successfully", "item": item.ToMap()})
}

func DeleteItemHandler(c *gin.Context) {
	item_id := c.Param("item_id")

	// handle role
	_, _, valid := helper.CheckUserRoleAndID(c, "admin")
	if !valid {
		return
	}

	// Pastikan item ada
	var item model.Item
	if err := config.DB.First(&item, item_id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Item not found"})
		return
	}

	// Periksa apakah item digunakan dalam transaksi
	var transactionCount int64
	if err := config.DB.Model(&model.Transaction{}).Where("item_id = ?", item_id).Count(&transactionCount).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to check item usage in transactions"})
		return
	}

	if transactionCount > 0 {
		c.JSON(400, gin.H{"error": "Cannot delete item: Item is used in transactions"})
		return
	}

	if err := config.DB.Unscoped().Delete(&item).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Item deleted successfully"})
}
