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

	var transaction []struct {
		ID       uint   `json:"transaction_id"`
		User     string `json:"user"`
		ItemName string `json:"item_name"`
		Quantity int    `json:"quantity"`
		Status   string `json:"status"`
	}

	query := config.DB.Table("transaction").
		Select("user.name AS user, transaction.id, transaction.quantity, transaction.status, item.name AS item_name").
		Joins("LEFT JOIN item ON item.id = transaction.item_id").
		Joins("LEFT JOIN user ON user.id = transaction.user_id")

	// Jika role adalah user, filter transaksi berdasarkan user_id
	if role == "user" {
		query = query.Where("transaction.user_id = ?", currentUserID)
	}

	if err := query.Scan(&transaction).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"transaction": transaction})
}

func UpdateTransactionHandler(c *gin.Context) {
	chart_id := c.Param("chart_id")

	// Ambil current_id dan role dari context (diset oleh middleware)
	currentUserID, exists := c.Get("current_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	role, roleExists := c.Get("role")
	if !roleExists || (role != "user" && chart_id != fmt.Sprint(currentUserID)) {
		c.JSON(403, gin.H{"error": "Forbidden: You can only access your own data"})
		return
	}

	// Pastikan transaction ada
	var transaction model.Transaction
	if err := config.DB.First(&transaction, chart_id).Error; err != nil {
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
	chartID := c.Param("chart_id")

	// Ambil current_id dan role dari context (diset oleh middleware)
	currentUserID, exists := c.Get("current_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	role, roleExists := c.Get("role")
	if !roleExists {
		c.JSON(403, gin.H{"error": "Forbidden: Role not found"})
		return
	}

	// Pastikan transaction ada
	var transaction model.Transaction
	if err := config.DB.First(&transaction, chartID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Transaction not found"})
		return
	}

	// Jika role adalah user, pastikan transaksi miliknya
	if role == "user" && transaction.UserID != currentUserID {
		c.JSON(403, gin.H{"error": "Forbidden: You can only delete your own transaction"})
		return
	}

	// Cek status transaksi
	if transaction.Status != "draft" {
		c.JSON(400, gin.H{"error": "Transaction can only be deleted if the status is 'draft'"})
		return
	}

	// Hapus transaction
	if err := config.DB.Unscoped().Delete(&transaction).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Transaction deleted successfully"})
}
