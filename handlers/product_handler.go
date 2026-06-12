package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"wms/config"
	"wms/models"
)

// CreateProduct menambahkan produk baru (admin only)
func CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	err := config.DB.QueryRow(
		`INSERT INTO products (name, sku, price, current_stock) VALUES ($1, $2, $3, 0) RETURNING id`,
		req.Name, req.SKU, req.Price,
	).Scan(&id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membuat produk, SKU mungkin sudah ada"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "name": req.Name, "sku": req.SKU, "price": req.Price, "current_stock": 0})
}

// GetProducts mengambil daftar semua produk beserta stok saat ini
func GetProducts(c *gin.Context) {
	rows, err := config.DB.Query(`SELECT id, name, sku, current_stock, price, created_at FROM products ORDER BY id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk"})
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.SKU, &p.CurrentStock, &p.Price, &p.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca data produk"})
			return
		}
		products = append(products, p)
	}

	c.JSON(http.StatusOK, products)
}

// GetProductTransactions mengambil history transaksi stok untuk produk tertentu
func GetProductTransactions(c *gin.Context) {
	productID := c.Param("id")

	rows, err := config.DB.Query(
		`SELECT id, product_id, type, quantity, note, created_by, created_at 
		 FROM stock_transactions WHERE product_id = $1 ORDER BY created_at DESC`,
		productID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil history transaksi"})
		return
	}
	defer rows.Close()

	var transactions []models.StockTransaction
	for rows.Next() {
		var t models.StockTransaction
		if err := rows.Scan(&t.ID, &t.ProductID, &t.Type, &t.Quantity, &t.Note, &t.CreatedBy, &t.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca data transaksi"})
			return
		}
		transactions = append(transactions, t)
	}

	c.JSON(http.StatusOK, transactions)
}
