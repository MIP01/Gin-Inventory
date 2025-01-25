package main

import (
	"Gin-Inventory/config"
	"Gin-Inventory/middleware"
	"Gin-Inventory/route"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Inisialisasi konfigurasi dan koneksi database
	config.InitConfig()

	// Inisialisasi router
	r := gin.Default()

	// Tambahkan Middleware CORS
	r.Use(middleware.CORSMiddleware())

	// Set up routes
	api := r.Group("/api/v1")
	route.SetupUserRoutes(api)
	route.SetupItemRoutes(api)

	// Informasi URL API
	log.Println("Server is running at http://localhost:8080")

	// Jalankan server pada port 8080
	r.Run(":8080")
}
