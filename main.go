package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"wms/config"
	"wms/routes"
)

func main() {
	// Load environment variables dari file .env
	if err := godotenv.Load(); err != nil {
		log.Println("Tidak menemukan file .env, menggunakan environment variable sistem")
	}

	// Koneksi ke database
	config.ConnectDB()

	// Auto migration: bikin tabel kalau belum ada
	config.RunMigrations()

	// Setup Gin router
	r := gin.Default()
	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
