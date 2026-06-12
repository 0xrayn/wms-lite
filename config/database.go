package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// schemaSQL berisi definisi tabel, dijalankan otomatis saat startup
// untuk membuat tabel kalau belum ada (idempotent karena pakai IF NOT EXISTS)
const schemaSQL = `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'staff',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    sku VARCHAR(100) UNIQUE NOT NULL,
    current_stock INT NOT NULL DEFAULT 0,
    price NUMERIC(12, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stock_transactions (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id),
    type VARCHAR(10) NOT NULL CHECK (type IN ('IN', 'OUT')),
    quantity INT NOT NULL CHECK (quantity > 0),
    note TEXT,
    created_by INT REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stock_transactions_product_id ON stock_transactions(product_id);
`

var DB *sql.DB

// ConnectDB membuka koneksi ke PostgreSQL menggunakan environment variables
func ConnectDB() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Gagal membuka koneksi database: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Gagal konek ke database: %v", err)
	}

	DB = db
	log.Println("Database connected successfully")
}

// RunMigrations menjalankan schemaSQL untuk membuat tabel kalau belum ada.
// Dipanggil sekali saat aplikasi start, jadi kita nggak perlu menjalankan
// schema.sql secara manual lewat psql.
func RunMigrations() {
	if _, err := DB.Exec(schemaSQL); err != nil {
		log.Fatalf("Gagal menjalankan migration: %v", err)
	}
	log.Println("Migration berhasil, semua tabel siap")
}
