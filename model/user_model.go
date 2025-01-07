package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name         string        `gorm:"size:100;not null"`
	Email        string        `gorm:"size:100;unique;not null"`
	Password     string        `gorm:"size:255;not null"`
	Role         string        `gorm:"size:50;not null" default:"user"`
	Transactions []Transaction `gorm:"foreignKey:UserID"`
}

func (u *User) TableName() string {
	return "user"
}

// Tambahkan metode ToMap untuk konversi user ke map
func (u *User) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"user_id": u.ID,
		"name":    u.Name,
		"email":   u.Email,
	}
}

// Fungsi untuk mengonversi slice User ke slice map
func UsersToMap(users []User) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, user := range users {
		result = append(result, user.ToMap())
	}
	return result
}

type Admin struct {
	gorm.Model
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"size:100;not null"`
	Email    string `gorm:"size:100;unique;not null"`
	Password string `gorm:"size:255;not null"`
	Role     string `gorm:"size:50;not null" default:"admin"`
}

func (u *Admin) TableName() string {
	return "admin"
}

// Tambahkan metode ToMap untuk konversi user ke map
func (u *Admin) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"admin_id": u.ID,
		"name":     u.Name,
		"email":    u.Email,
	}
}

// Fungsi untuk mengonversi slice User ke slice map
func AdminsToMap(admins []Admin) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, admin := range admins {
		result = append(result, admin.ToMap())
	}
	return result
}
