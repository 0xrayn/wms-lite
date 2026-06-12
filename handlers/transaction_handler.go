package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"wms/config"
	"wms/models"
)

// CreateTransaction mencatat transaksi stok (IN/OUT) dan memperbarui current_stock
// secara atomic menggunakan database transaction + row locking.
func CreateTransaction(c *gin.Context) {
	var req models.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	// Mulai database transaction
	tx, err := config.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memulai transaksi"})
		return
	}
	// Rollback otomatis kalau terjadi error di tengah jalan dan belum di-commit
	defer tx.Rollback()

	// Lock baris produk ini supaya request lain yang menyentuh produk yang sama
	// harus menunggu sampai transaksi ini selesai (mencegah race condition)
	var currentStock int
	err = tx.QueryRow(
		`SELECT current_stock FROM products WHERE id = $1 FOR UPDATE`,
		req.ProductID,
	).Scan(&currentStock)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk"})
		return
	}

	// Hitung stok baru berdasarkan tipe transaksi
	var newStock int
	if req.Type == "IN" {
		newStock = currentStock + req.Quantity
	} else { // OUT
		if currentStock < req.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":         "Stok tidak cukup",
				"current_stock": currentStock,
				"requested":     req.Quantity,
			})
			return
		}
		newStock = currentStock - req.Quantity
	}

	// Update stok produk
	_, err = tx.Exec(
		`UPDATE products SET current_stock = $1 WHERE id = $2`,
		newStock, req.ProductID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update stok produk"})
		return
	}

	// Catat transaksi ke history
	var transactionID int
	err = tx.QueryRow(
		`INSERT INTO stock_transactions (product_id, type, quantity, note, created_by) 
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		req.ProductID, req.Type, req.Quantity, req.Note, userID,
	).Scan(&transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencatat transaksi"})
		return
	}

	// Kalau semua berhasil, commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan transaksi"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"transaction_id": transactionID,
		"product_id":     req.ProductID,
		"type":           req.Type,
		"quantity":       req.Quantity,
		"new_stock":      newStock,
	})
}
