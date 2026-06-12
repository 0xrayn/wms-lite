package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"wms/config"
)

type StockSummary struct {
	ProductID    int     `json:"product_id"`
	Name         string  `json:"name"`
	SKU          string  `json:"sku"`
	CurrentStock int     `json:"current_stock"`
	TotalIn      int     `json:"total_in"`
	TotalOut     int     `json:"total_out"`
	StockValue   float64 `json:"stock_value"`
}

// GetStockSummary mengambil ringkasan stok semua produk, termasuk total IN/OUT
func GetStockSummary(c *gin.Context) {
	rows, err := config.DB.Query(`
		SELECT 
			p.id, p.name, p.sku, p.current_stock, p.price,
			COALESCE(SUM(CASE WHEN st.type = 'IN' THEN st.quantity ELSE 0 END), 0) AS total_in,
			COALESCE(SUM(CASE WHEN st.type = 'OUT' THEN st.quantity ELSE 0 END), 0) AS total_out
		FROM products p
		LEFT JOIN stock_transactions st ON st.product_id = p.id
		GROUP BY p.id, p.name, p.sku, p.current_stock, p.price
		ORDER BY p.id
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil ringkasan stok"})
		return
	}
	defer rows.Close()

	var summaries []StockSummary
	for rows.Next() {
		var s StockSummary
		var price float64
		if err := rows.Scan(&s.ProductID, &s.Name, &s.SKU, &s.CurrentStock, &price, &s.TotalIn, &s.TotalOut); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca data ringkasan"})
			return
		}
		s.StockValue = price * float64(s.CurrentStock)
		summaries = append(summaries, s)
	}

	c.JSON(http.StatusOK, summaries)
}
