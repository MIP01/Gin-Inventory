package controller

import (
	"Gin-Inventory/config"
	"Gin-Inventory/helper"
	"Gin-Inventory/middleware"
	"Gin-Inventory/model"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateDetailHandler(c *gin.Context) {
	// handle role
	currentUserID, _, valid := helper.CheckUserRoleAndID(c, "user")
	if !valid {
		return
	}

	// Memvalidasi input dengan Middleware ValidateInput.
	detailData, valid := helper.ValidationHelper(c, middleware.DetailSchema{})
	if !valid {
		return
	}

	// Cari transaksi milik current_user dengan status 'draft'
	var transactions []model.Transaction
	if err := config.DB.Where("user_id = ? AND status = ?", currentUserID, "draft").Find(&transactions).Error; err != nil || len(transactions) == 0 {
		c.JSON(404, gin.H{"error": "No draft transaction found for the current user"})
		return
	}

	// Konversi string "Out" dan "Entry" ke time.Time
	var outTime, entryTime time.Time
	var err error
	if detailData.Out != "" {
		outTime, err = time.Parse("2006-01-02", detailData.Out)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid date format for Out"})
			return
		}
	}

	if detailData.Entry != "" {
		entryTime, err = time.Parse("2006-01-02", detailData.Entry)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid date format for Entry"})
			return
		}
	}

	// Buat detail baru
	autoGeneratedCode := fmt.Sprintf("ivt%s%d", time.Now().Format("0405"), currentUserID)
	newDetail := model.Detail{
		Code:   autoGeneratedCode,
		Out:    outTime,
		Entry:  entryTime,
		Status: "pending",
	}

	if err := config.DB.Create(&newDetail).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Loop melalui transaksi untuk memperbarui masing-masing
	var updatedTransactions []model.Transaction
	for _, trx := range transactions {
		// Perbarui setiap transaksi dengan DetailID baru
		trx.DetailID = &newDetail.ID
		trx.Status = "pending" // Ubah status
		if err := config.DB.Save(&trx).Error; err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to update transaction ID %d", trx.ID)})
			return
		}
		updatedTransactions = append(updatedTransactions, trx)
	}

	// Kirim respons setelah semua transaksi diproses
	c.JSON(201, gin.H{
		"message":      "Detail created successfully",
		"detail":       newDetail.ToMap(),
		"transactions": updatedTransactions,
	})
}

func GetAllDetailHandler(c *gin.Context) {
	// handle role
	currentUserID, role, valid := helper.CheckUserRoleAndID(c, "user", "admin")
	if !valid {
		return
	}

	var detail []struct {
		ID        uint      `json:"detail_id"`
		Code      string    `json:"code"`
		User      string    `json:"user"`
		Out       time.Time `json:"out"`
		Entry     time.Time `json:"entry"`
		Status    string    `json:"status"`
		Quantity  int       `json:"quantity"`
		ItemName  string    `json:"item_name"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`
	}

	query := config.DB.Table("detail").
		Select(`user.name AS user, detail.id, detail.code, detail.out, detail.entry, detail.status, detail.created_at,
		detail.updated_at, transaction.quantity, item.name AS item_name`).
		Joins("LEFT JOIN transaction ON transaction.detail_id = detail.id").
		Joins("LEFT JOIN item ON item.id = transaction.item_id").
		Joins("LEFT JOIN user ON user.id = transaction.user_id")

	if role == "user" {
		query = query.Where("transaction.user_id = ?", currentUserID)
	} else if role != "admin" {
		c.JSON(403, gin.H{"error": "Forbidden: Invalid role"})
		return
	}

	if err := query.Scan(&detail).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Periksa apakah detail ditemukan
	if len(detail) == 0 {
		c.JSON(404, gin.H{"error": "Detail not found"})
		return
	}

	c.JSON(200, gin.H{"detail": detail})
}

func GetDetailHandler(c *gin.Context) {
	detail_id := c.Param("detail_id")

	// handle role
	currentUserID, role, valid := helper.CheckUserRoleAndID(c, "user", "admin")
	if !valid {
		return
	}

	// Jika role adalah user, pastikan detail miliknya
	if role == "user" {
		var transaction model.Transaction
		if err := config.DB.Where("detail_id = ? AND user_id = ?", detail_id, currentUserID).First(&transaction).Error; err != nil {
			c.JSON(403, gin.H{"error": "Forbidden: You can only delete your own detail"})
			return
		}
	}

	// Definisikan struktur untuk hasil query
	var detail struct {
		Code     string    `json:"code"`
		User     string    `json:"user"`
		Out      time.Time `json:"out"`
		Entry    time.Time `json:"entry"`
		Status   string    `json:"status"`
		Quantity int       `json:"quantity"`
		ItemName string    `json:"item_name"`
	}

	// Query dengan join untuk mendapatkan data yang dibutuhkan
	err := config.DB.Table("detail").
		Select("user.name AS user, detail.code, detail.out, detail.entry, detail.status, transaction.quantity, item.name AS item_name").
		Joins("LEFT JOIN transaction ON transaction.detail_id = detail.id").
		Joins("LEFT JOIN item ON item.id = transaction.item_id").
		Joins("LEFT JOIN user ON user.id = transaction.user_id").
		Where("detail.id = ?", detail_id).
		Scan(&detail).Error

	// Periksa apakah detail ditemukan
	if err != nil || (detail.Code == "" && detail.Status == "" && detail.ItemName == "") {
		c.JSON(404, gin.H{"error": "Detail not found"})
		return
	}

	c.JSON(200, gin.H{"detail": detail})
}

func UpdateDetailHandler(c *gin.Context) {
	detail_id := c.Param("detail_id")

	// handle role
	currentUserID, role, valid := helper.CheckUserRoleAndID(c, "user", "admin")
	if !valid {
		return
	}

	// Jika role adalah user, pastikan detail miliknya
	if role == "user" {
		var transaction model.Transaction
		if err := config.DB.Where("detail_id = ? AND user_id = ?", detail_id, currentUserID).First(&transaction).Error; err != nil {
			c.JSON(403, gin.H{"error": "Forbidden: You can only delete your own detail"})
			return
		}
	}

	// Pastikan detail ada
	var detail model.Detail
	if err := config.DB.Preload("Transactions").First(&detail, detail_id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Detail not found"})
		return
	}

	// Memvalidasi input dengan Middleware ValidateInput.
	updatedData, valid := helper.ValidationHelper(c, middleware.DetailSchema{})
	if !valid {
		return
	}

	// Menyimpan status sebelumnya
	previousStatus := detail.Status

	// Jika user dan status adalah pending, izinkan perubahan Out dan Entry
	if role == "user" && detail.Status == "pending" {
		if updatedData.Out != "" {
			detail.Out, _ = time.Parse("2006-01-02", updatedData.Out)
		}
		if updatedData.Entry != "" {
			detail.Entry, _ = time.Parse("2006-01-02", updatedData.Entry)
		}
	}

	// Jika admin, hanya bisa mengubah status
	if role == "admin" {
		if updatedData.Status != "" {
			detail.Status = updatedData.Status
		}

		// Jika status berubah dari 'pending' ke 'loaned', kurangi quantity dari stok item
		if previousStatus == "pending" && detail.Status == "loaned" {
			for _, transaction := range detail.Transactions {
				var item model.Item
				if err := config.DB.First(&item, transaction.ItemID).Error; err != nil {
					c.JSON(404, gin.H{"error": fmt.Sprintf("Item with ID %d not found", transaction.ItemID)})
					return
				}

				// Kurangi stok jika status berubah dari pending ke loaned
				if item.Stock >= transaction.Quantity {
					item.Stock -= transaction.Quantity
					if err := config.DB.Save(&item).Error; err != nil {
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to update item stock for item ID %d", item.ID)})
						return
					}

					// Update status transaksi menjadi finish
					transaction.Status = "finish"
					if err := config.DB.Save(&transaction).Error; err != nil {
						c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to update transaction ID %d", transaction.ID)})
						return
					}
				} else {
					c.JSON(400, gin.H{"error": fmt.Sprintf("Not enough stock available for item %s", item.Name)})
					return
				}
			}
		}

		// Jika status berubah dari 'loaned' ke 'return' atau 'pending', kembalikan quantity ke stok
		if previousStatus == "loaned" && (detail.Status == "return" || detail.Status == "pending" || detail.Status == "rejected") {
			for _, transaction := range detail.Transactions {
				var item model.Item
				if err := config.DB.First(&item, transaction.ItemID).Error; err != nil {
					c.JSON(404, gin.H{"error": fmt.Sprintf("Item with ID %d not found", transaction.ItemID)})
					return
				}

				// Kembalikan stok jika status berubah menjadi 'return' atau 'pending'
				item.Stock += transaction.Quantity
				if err := config.DB.Save(&item).Error; err != nil {
					c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to update item stock for item ID %d", item.ID)})
					return
				}
			}
		}

		// Hanya mengizinkan perubahan ke pending
		if previousStatus == "rejected" && detail.Status != "pending" {
			c.JSON(400, gin.H{"error": "rejected status can only be changed to pending"})
			return
		}
	}

	// Jika status berubah dari 'pending' ke 'rejected', tidak ada perubahan stok
	if previousStatus == "pending" && detail.Status == "rejected" {
		// Tidak ada tindakan yang diperlukan, hanya mengubah status
	}

	// Simpan perubahan
	if err := config.DB.Save(&detail).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Detail updated successfully", "detail": detail.ToMap()})
}

func DeleteDetailHandler(c *gin.Context) {
	detailID := c.Param("detail_id")

	// handle role
	currentUserID, role, valid := helper.CheckUserRoleAndID(c, "user", "admin")
	if !valid {
		return
	}

	// Pastikan detail ada
	var detail model.Detail
	if err := config.DB.Preload("Transactions").First(&detail, detailID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Detail not found"})
		return
	}

	// Jika role adalah user, pastikan detail miliknya
	if role == "user" {
		var transaction model.Transaction
		if err := config.DB.Where("detail_id = ? AND user_id = ?", detailID, currentUserID).First(&transaction).Error; err != nil {
			c.JSON(403, gin.H{"error": "Forbidden: You can only delete your own detail"})
			return
		}
	}

	// Periksa status detail
	if detail.Status != "pending" && detail.Status != "rejected" {
		c.JSON(400, gin.H{"error": "Cannot delete detail: Detail status must be pending or rejected"})
		return
	}

	// Hapus transaksi yang terhubung dengan detail_id
	if err := config.DB.Unscoped().Where("detail_id = ?", detailID).Delete(&model.Transaction{}).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete related transactions"})
		return
	}

	if err := config.DB.Unscoped().Delete(&detail).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Detail and related transactions deleted successfully"})
}
