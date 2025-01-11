package controller

import (
	"gin_iventory/config"
	"gin_iventory/middleware"
	"gin_iventory/model"

	"github.com/gin-gonic/gin"
)

func CreateTransactionHandler(c *gin.Context) {
	currentUserID, currentUserExists := c.Get("current_id")
	role, roleExists := c.Get("role")
	if !currentUserExists || !roleExists || role != "user" {
		c.JSON(403, gin.H{"error": "Unauthorized"})
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
		UserID:   currentUserID.(uint),
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
	currentUserID, currentUserExists := c.Get("current_id")
	role, roleExists := c.Get("role")

	// Periksa jika role atau currentUserID tidak ditemukan
	if !currentUserExists || !roleExists || currentUserID == nil {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var transaction []struct {
		ID        uint   `json:"transaction_id"`
		User      string `json:"user"`
		ItemName  string `json:"item_name"`
		Quantity  int    `json:"quantity"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	query := config.DB.Table("transaction").
		Select(`user.name AS user, transaction.id, transaction.quantity, transaction.status, item.name AS item_name, transaction.created_at,
		transaction.updated_at`).
		Joins("LEFT JOIN item ON item.id = transaction.item_id").
		Joins("LEFT JOIN user ON user.id = transaction.user_id")

	// Jika role adalah user, filter transaksi berdasarkan user_id
	if role == "user" {
		query = query.Where("transaction.user_id = ?", currentUserID)
	} else if role != "admin" {
		c.JSON(403, gin.H{"error": "Forbidden: Invalid role"})
		return
	}

	if err := query.Scan(&transaction).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"transaction": transaction})
}

func UpdateTransactionHandler(c *gin.Context) {
	chart_id := c.Param("chart_id")

	// Ambil current_id dan role dari context
	currentUserID, currentUserExists := c.Get("current_id")
	role, roleExists := c.Get("role")

	// Periksa jika role atau currentUserID tidak ditemukan
	if !currentUserExists || !roleExists || currentUserID == nil {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// Pastikan transaction ada
	var transaction model.Transaction
	if err := config.DB.First(&transaction, chart_id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Transaction not found"})
		return
	}

	// Jika role adalah user, pastikan transaksi miliknya
	if role == "user" {
		// Pastikan transaksi milik pengguna saat ini
		if err := config.DB.Where("id = ? AND user_id = ?", chart_id, currentUserID).First(&transaction).Error; err != nil {
			c.JSON(403, gin.H{"error": "Forbidden: You can only update your own transaction"})
			return
		}
	} else if role != "admin" {
		c.JSON(403, gin.H{"error": "Forbidden: Invalid role"})
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

	// Ambil current_id dan role dari context
	currentUserID, currentUserExists := c.Get("current_id")
	role, roleExists := c.Get("role")

	// Periksa jika role atau currentUserID tidak ditemukan
	if !currentUserExists || !roleExists || currentUserID == nil {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// Pastikan transaction ada
	var transaction model.Transaction
	if err := config.DB.First(&transaction, chartID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Transaction not found"})
		return
	}

	// Jika role adalah user, pastikan transaksi miliknya
	if role == "user" {
		// Pastikan transaksi milik pengguna saat ini
		if err := config.DB.Where("id = ? AND user_id = ?", chartID, currentUserID).First(&transaction).Error; err != nil {
			c.JSON(403, gin.H{"error": "Forbidden: You can only update your own transaction"})
			return
		}
	} else if role != "admin" {
		c.JSON(403, gin.H{"error": "Forbidden: Invalid role"})
		return
	}

	// Cek status transaksi
	if transaction.Status != "draft" {
		c.JSON(400, gin.H{"error": "Transaction can only be deleted if the status is 'draft'"})
		return
	}

	if err := config.DB.Unscoped().Delete(&transaction).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Transaction deleted successfully"})
}
