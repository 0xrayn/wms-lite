package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"wms/config"
	"wms/models"
)

// parsePagination membaca query param page & limit, dengan default dan batas aman
func parsePagination(c *gin.Context) (page, limit, offset int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err = strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset = (page - 1) * limit
	return page, limit, offset
}

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

// GetProducts mengambil daftar produk dengan pagination dan optional search by name/sku
// Query params: page, limit, search
func GetProducts(c *gin.Context) {
	page, limit, offset := parsePagination(c)
	search := c.Query("search")

	var total int
	var err error
	var rows *sql.Rows

	if search != "" {
		pattern := "%" + search + "%"
		err = config.DB.QueryRow(
			`SELECT COUNT(*) FROM products WHERE name ILIKE $1 OR sku ILIKE $1`, pattern,
		).Scan(&total)
		if err == nil {
			rows, err = config.DB.Query(
				`SELECT id, name, sku, current_stock, price, created_at FROM products 
				 WHERE name ILIKE $1 OR sku ILIKE $1 
				 ORDER BY id LIMIT $2 OFFSET $3`,
				pattern, limit, offset,
			)
		}
	} else {
		err = config.DB.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&total)
		if err == nil {
			rows, err = config.DB.Query(
				`SELECT id, name, sku, current_stock, price, created_at FROM products 
				 ORDER BY id LIMIT $1 OFFSET $2`,
				limit, offset,
			)
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk"})
		return
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.SKU, &p.CurrentStock, &p.Price, &p.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca data produk"})
			return
		}
		products = append(products, p)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetLowStockProducts mengambil produk dengan current_stock di bawah threshold
// Query params: threshold (default 10), page, limit
func GetLowStockProducts(c *gin.Context) {
	page, limit, offset := parsePagination(c)

	threshold, err := strconv.Atoi(c.DefaultQuery("threshold", "10"))
	if err != nil || threshold < 0 {
		threshold = 10
	}

	var total int
	if err := config.DB.QueryRow(
		`SELECT COUNT(*) FROM products WHERE current_stock < $1`, threshold,
	).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung produk low stock"})
		return
	}

	rows, err := config.DB.Query(
		`SELECT id, name, sku, current_stock, price, created_at FROM products 
		 WHERE current_stock < $1 ORDER BY current_stock ASC LIMIT $2 OFFSET $3`,
		threshold, limit, offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk low stock"})
		return
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.SKU, &p.CurrentStock, &p.Price, &p.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca data produk"})
			return
		}
		products = append(products, p)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      products,
		"threshold": threshold,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetProductTransactions mengambil history transaksi stok untuk produk tertentu, dengan pagination
// Query params: page, limit
func GetProductTransactions(c *gin.Context) {
	productID := c.Param("id")
	page, limit, offset := parsePagination(c)

	var total int
	if err := config.DB.QueryRow(
		`SELECT COUNT(*) FROM stock_transactions WHERE product_id = $1`, productID,
	).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung total transaksi"})
		return
	}

	rows, err := config.DB.Query(
		`SELECT id, product_id, type, quantity, note, created_by, created_at 
		 FROM stock_transactions WHERE product_id = $1 
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		productID, limit, offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil history transaksi"})
		return
	}
	defer rows.Close()

	transactions := []models.StockTransaction{}
	for rows.Next() {
		var t models.StockTransaction
		if err := rows.Scan(&t.ID, &t.ProductID, &t.Type, &t.Quantity, &t.Note, &t.CreatedBy, &t.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca data transaksi"})
			return
		}
		transactions = append(transactions, t)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": transactions,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
