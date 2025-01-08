package controller

import (
	"fmt"
	"gin_iventory/config"
	"gin_iventory/middleware"
	"gin_iventory/model"

	"github.com/gin-gonic/gin"
)

func CreateTransactionHandler(c *gin.Context) {
	role, roleExists := c.Get("role")
	if !roleExists || role != "user" {
		c.JSON(403, gin.H{"error": "Unauthorized"})
		return
	}

	currentID, exists := c.Get("current_id")
	if !exists {
		c.JSON(500, gin.H{"error": "Failed to get user ID from context"})
		return
	}

	var transactionData middleware.TransactionSchema
	if err := c.ShouldBindJSON(&transactionData); err != nil {
		errors := middleware.FormatValidationErrors(err)
		c.JSON(400, gin.H{"errors": errors})
		return
	}

	validationErrors := middleware.ValidateInput(transactionData)
	if validationErrors != nil {
		c.JSON(400, gin.H{"errors": validationErrors})
		return
	}

	newTransaction := model.Transaction{
		UserID:   currentID.(uint),
		ItemID:   transactionData.ItemID,
		Quantity: transactionData.Quantity,
		Status:   "draft",
	}

	if err := config.DB.Create(&newTransaction).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Transaction created successfully", "transaction": newTransaction.ToMap()})
}

func GetTransactionsHandler(c *gin.Context) {
	// Ambil current_id dan role dari context (diset oleh middleware)
	currentUserID, exists := c.Get("current_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	role, roleExists := c.Get("role")
	if !roleExists {
		c.JSON(403, gin.H{"error": "Forbidden: Role is not specified"})
		return
	}

	var transactions []model.Transaction

	// Query dengan Preload
	query := config.DB.Preload("User").Preload("Detail").Preload("Item")

	// Jika role adalah admin, ambil semua transaksi
	if role == "admin" {
		if err := query.Find(&transactions).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	} else if role == "user" {
		// Jika role adalah user, ambil transaksi milik user tersebut
		if err := query.Where("user_id = ?", currentUserID).Find(&transactions).Error; err != nil {
			c.JSON(404, gin.H{"error": "Transaction not found"})
			return
		}
	} else {
		// Role tidak dikenali
		c.JSON(403, gin.H{"error": "Forbidden: Invalid role"})
		return
	}

	c.JSON(200, gin.H{"transactions": model.TransactionsToMap(transactions)})
}

func UpdateTransactionHandler(c *gin.Context) {
	id := c.Param("id")

	// Ambil current_id dan role dari context (diset oleh middleware)
	currentUserID, exists := c.Get("current_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	role, roleExists := c.Get("role")
	if !roleExists || (role != "user" && id != fmt.Sprint(currentUserID)) {
		c.JSON(403, gin.H{"error": "Forbidden: You can only access your own data"})
		return
	}

	// Pastikan transaction ada
	var transaction model.Transaction
	if err := config.DB.First(&transaction, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Transaction not found"})
		return
	}

	// Memvalidasi input dengan Middleware ValidateInput.
	var updatedData middleware.TransactionSchema
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

	// Cek status transaksi
	if transaction.Status != "draft" {
		c.JSON(400, gin.H{"error": "Transaction can only be updated if the status is 'draft'"})
		return
	}

	// Perbarui field yang diberikan
	if updatedData.ItemID != 0 {
		transaction.ItemID = updatedData.ItemID
	}
	if updatedData.Quantity != 0 {
		transaction.Quantity = updatedData.Quantity
	}
	if err := config.DB.Save(&transaction).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Transaction updated successfully", "transaction": transaction.ToMap()})
}

func DeleteTransactionHandler(c *gin.Context) {
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

	// Pastikan transaction ada
	var transaction model.Transaction
	if err := config.DB.First(&transaction, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Transaction not found"})
		return
	}

	// Cek status transaksi
	if transaction.Status != "draft" {
		c.JSON(400, gin.H{"error": "Transaction can only be deleted if the status is 'draft'"})
		return
	}

	// Hapus transaction
	if err := config.DB.Delete(&transaction).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Transaction deleted successfully"})
}
