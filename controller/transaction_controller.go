package controller

import (
	"Gin-Inventory/config"
	"Gin-Inventory/helper"
	"Gin-Inventory/middleware"
	"Gin-Inventory/model"

	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func CreateTransactionHandler(c *gin.Context) {
	// handle role
	currentUserID, _, valid := helper.CheckUserRoleAndID(c, "user")
	if !valid {
		return
	}

	// Memvalidasi input dengan Middleware ValidateInput.
	transactionData, valid := helper.ValidationHelper(c, middleware.TransactionSchema{})
	if !valid {
		return
	}

	// Cek stok item
	var item model.Item
	if err := config.DB.First(&item, transactionData.ItemID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Item not found"})
		return
	}

	// Cek apakah transaksi dengan UserID dan ItemID sudah ada
	var existingTransaction model.Transaction
	if err := config.DB.Where("user_id = ? AND item_id = ?", currentUserID, transactionData.ItemID).First(&existingTransaction).Error; err == nil {
		// Jika transaksi ditemukan, periksa statusnya
		if existingTransaction.Status == "draft" {
			// Hitung total quantity yang akan dimasukkan
			totalQuantity := existingTransaction.Quantity + transactionData.Quantity

			// Logging untuk debugging
			log.Printf("Checking stock for item %d: current stock %d, requested quantity %d", item.ID, item.Stock, totalQuantity)

			// Validasi stok
			if totalQuantity > item.Stock {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Not enough stock available for item %s. Requested: %d, Available: %d", item.Name, totalQuantity, item.Stock)})
				return
			}

			// Jika status "draft" dan stok mencukupi, tambahkan kuantitas
			existingTransaction.Quantity = totalQuantity
			if err := config.DB.Save(&existingTransaction).Error; err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"message": "Transaction updated successfully", "transaction": existingTransaction.ToMap()})
			return
		}
	}

	// Logging untuk debugging
	log.Printf("Creating new transaction for item %d: requested quantity %d, available stock %d", item.ID, transactionData.Quantity, item.Stock)

	// Validasi stok untuk transaksi baru
	if transactionData.Quantity > item.Stock {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Not enough stock available for item %s. Requested: %d, Available: %d", item.Name, transactionData.Quantity, item.Stock)})
		return
	}

	newTransaction := model.Transaction{
		UserID:   currentUserID,
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
	// handle role
	currentUserID, role, valid := helper.CheckUserRoleAndID(c, "user", "admin")
	if !valid {
		return
	}

	var transaction []struct {
		ID        uint   `json:"transaction_id"`
		ItemID    uint   `json:"item_id"`
		User      string `json:"user"`
		ItemName  string `json:"item_name"`
		Stock     int    `json:"stock"`
		Quantity  int    `json:"quantity"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	query := config.DB.Table("transaction").
		Select(`user.name AS user, transaction.id, transaction.quantity, transaction.status, item.id AS item_id, item.name AS item_name, item.stock AS stock, transaction.created_at,
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

	// Periksa apakah detail ditemukan
	if len(transaction) == 0 {
		c.JSON(404, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(200, gin.H{"transaction": transaction})
}

func UpdateTransactionHandler(c *gin.Context) {
	chart_id := c.Param("chart_id")

	// handle role
	currentUserID, role, valid := helper.CheckUserRoleAndID(c, "user", "admin")
	if !valid {
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
	}

	// Memvalidasi input dengan Middleware ValidateInput.
	updatedData, valid := helper.ValidationHelper(c, middleware.TransactionSchema{})
	if !valid {
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

	// handle role
	currentUserID, role, valid := helper.CheckUserRoleAndID(c, "user", "admin")
	if !valid {
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
