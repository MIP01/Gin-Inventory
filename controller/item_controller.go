package controller

import (
	"gin_iventory/config"
	"gin_iventory/middleware"
	"gin_iventory/model"

	"github.com/gin-gonic/gin"
)

func CreateItemHandler(c *gin.Context) {
	role, roleExists := c.Get("role")
	if !roleExists || role != "admin" {
		c.JSON(403, gin.H{"error": "Unauthorized"})
		return
	}

	var itemData middleware.ItemSchema
	if err := c.ShouldBindJSON(&itemData); err != nil {
		errors := middleware.FormatValidationErrors(err)
		c.JSON(400, gin.H{"errors": errors})
		return
	}

	validationErrors := middleware.ValidateInput(itemData)
	if validationErrors != nil {
		c.JSON(400, gin.H{"errors": validationErrors})
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
	// Ambil ID item dari parameter URL
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

	role, roleExists := c.Get("role")
	if !roleExists || role != "admin" {
		c.JSON(403, gin.H{"error": "Unauthorized"})
		return
	}

	// Pastikan item ada
	var item model.Item
	if err := config.DB.First(&item, item_id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Item not found"})
		return
	}

	// Memvalidasi input dengan Middleware ValidateInput.
	var updatedData middleware.ItemSchema
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

	role, roleExists := c.Get("role")
	if !roleExists || role != "admin" {
		c.JSON(403, gin.H{"error": "Unauthorized"})
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

	// Hapus item
	if err := config.DB.Delete(&item).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Item deleted successfully"})
}
