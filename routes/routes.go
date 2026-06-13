package routes

import (
	"github.com/gin-gonic/gin"

	"wms/handlers"
	"wms/middleware"
)

// SetupRoutes mendaftarkan semua endpoint API
func SetupRoutes(r *gin.Engine) {
	// Public routes
	r.POST("/auth/register", handlers.Register)
	r.POST("/auth/login", handlers.Login)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Protected routes (butuh login)
	auth := r.Group("/")
	auth.Use(middleware.AuthRequired())
	{
		auth.GET("/products", handlers.GetProducts)
		auth.GET("/products/low-stock", handlers.GetLowStockProducts)
		auth.GET("/products/:id/transactions", handlers.GetProductTransactions)
		auth.POST("/transactions", handlers.CreateTransaction)
		auth.GET("/reports/stock-summary", handlers.GetStockSummary)

		// Admin only routes
		admin := auth.Group("/")
		admin.Use(middleware.AdminOnly())
		{
			admin.POST("/products", handlers.CreateProduct)
		}
	}
}
