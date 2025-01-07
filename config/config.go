package config

import (
	"log"
	"os"
	"time"

	"gin_iventory/model"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var JWTSecret string

func InitConfig() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	JWTSecret = os.Getenv("JWT_SECRET")

	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// AutoMigrate models
	if err := db.AutoMigrate(&model.User{}, &model.Admin{}, &model.Item{}, &model.Detail{}, &model.Transaction{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	DB = db
	log.Println("Database connected successfully!")
}

func JWTExpireDuration() time.Duration {
	return time.Hour * 1
}
