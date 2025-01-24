package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Item struct {
	gorm.Model
	Name        string        `gorm:"size:100;unique;not null"`
	Stock       int           `gorm:"not null"`
	Transaction []Transaction `gorm:"foreignKey:ItemID"`
}

func (u *Item) TableName() string {
	return "item"
}

// Tambahkan metode ToMap untuk konversi user ke map
func (u *Item) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"item_id":    u.ID,
		"name":       u.Name,
		"stock":      u.Stock,
		"created_at": u.CreatedAt.Format(time.RFC3339),
		"updated_at": u.UpdatedAt.Format(time.RFC3339),
	}
}

// Fungsi untuk mengonversi slice User ke slice map
func ItemsToMap(items []Item) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, item := range items {
		result = append(result, item.ToMap())
	}
	return result
}

type Detail struct {
	gorm.Model
	Code         string        `gorm:"size:100;not null"`
	Out          time.Time     `gorm:"null"`
	Entry        time.Time     `gorm:"null"`
	Status       string        `gorm:"size:50;not null;default:'pending'"`
	Transactions []Transaction `gorm:"foreignKey:DetailID;constraint:OnDelete:CASCADE;"`
}

// BeforeSave hook untuk validasi Status
func (t *Detail) BeforeSave(tx *gorm.DB) error {
	allowedStatuses := []string{"pending", "loaned", "return", "rejected"}
	for _, allowedStatus := range allowedStatuses {
		if t.Status == allowedStatus {
			return nil
		}
	}
	return fmt.Errorf("invalid status: %s, allowed values are: pending, loaned, return, rejected", t.Status)
}

func (u *Detail) TableName() string {
	return "Detail"
}

// Tambahkan metode ToMap untuk konversi user ke map
func (u *Detail) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"detail_id":  u.ID,
		"code":       u.Code,
		"out":        u.Out,
		"entry":      u.Entry,
		"status":     u.Status,
		"created_at": u.CreatedAt.Format(time.RFC3339),
		"updated_at": u.UpdatedAt.Format(time.RFC3339),
	}
}

// Fungsi untuk mengonversi slice User ke slice map
func DetailsToMap(details []Detail) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, detail := range details {
		result = append(result, detail.ToMap())
	}
	return result
}

type Transaction struct {
	gorm.Model
	UserID   uint   `gorm:"not null"`
	DetailID *uint  `gorm:"null"`
	ItemID   uint   `gorm:"not null"`
	Quantity int    `gorm:"not null"`
	Status   string `gorm:"size:50;not null;default:'draft'"`
	User     User   `gorm:"foreignKey:UserID"`
	Detail   Detail `gorm:"foreignKey:DetailID;constraint:OnDelete:CASCADE;"`
	Item     Item   `gorm:"foreignKey:ItemID;constraint:OnDelete:CASCADE;"`
}

// BeforeSave hook untuk validasi Status
func (t *Transaction) BeforeSave(tx *gorm.DB) error {
	allowedStatuses := []string{"draft", "pending", "finish"}
	for _, allowedStatus := range allowedStatuses {
		if t.Status == allowedStatus {
			return nil
		}
	}
	return fmt.Errorf("invalid status: %s, allowed values are: draft, pending, finish", t.Status)
}

func (u *Transaction) TableName() string {
	return "transaction"
}

// Tambahkan metode ToMap untuk konversi user ke map
func (u *Transaction) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"transaction_id": u.ID,
		"user_id":        u.UserID,
		"detail_id":      u.DetailID,
		"item_id":        u.ItemID,
		"quantity":       u.Quantity,
		"status":         u.Status,
		"created_at":     u.CreatedAt.Format(time.RFC3339),
		"updated_at":     u.UpdatedAt.Format(time.RFC3339),
	}
}

// Fungsi untuk mengonversi slice User ke slice map
func TransactionsToMap(transactions []Transaction) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, transaction := range transactions {
		result = append(result, transaction.ToMap())
	}
	return result
}
