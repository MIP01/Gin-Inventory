package model

import (
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
		"item_id": u.ID,
		"name":    u.Name,
		"stock":   u.Stock,
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
	Transactions []Transaction `gorm:"foreignKey:DetailID"`
}

func (u *Detail) TableName() string {
	return "Detail"
}

// Tambahkan metode ToMap untuk konversi user ke map
func (u *Detail) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"detail_id": u.ID,
		"code":      u.Code,
		"out":       u.Out,
		"entry":     u.Entry,
		"status":    u.Status,
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
	DetailID uint   `gorm:"null"`
	ItemID   uint   `gorm:"not null"`
	Quantity int    `gorm:"not null"`
	Status   string `gorm:"size:50;not null;default:'draft'"`
	User     User   `gorm:"foreignKey:UserID"`
	Detail   Detail `gorm:"foreignKey:DetailID"`
	Item     Item   `gorm:"foreignKey:ItemID"`
}

func (u *Transaction) TableName() string {
	return "transaction"
}

// Tambahkan metode ToMap untuk konversi user ke map
func (u *Transaction) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"user_id":     u.UserID,
		"detail_id":   u.DetailID,
		"item_id":     u.ItemID,
		"user_name":   u.User.Name,
		"detail_code": u.Detail.Code,
		"item_name":   u.Item.Name,
		"quantity":    u.Quantity,
		"status":      u.Status,
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
