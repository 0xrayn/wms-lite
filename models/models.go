package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Product struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	SKU          string    `json:"sku"`
	CurrentStock int       `json:"current_stock"`
	Price        float64   `json:"price"`
	CreatedAt    time.Time `json:"created_at"`
}

type StockTransaction struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	Type      string    `json:"type"` // "IN" or "OUT"
	Quantity  int       `json:"quantity"`
	Note      string    `json:"note"`
	CreatedBy int       `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// Request payloads

type CreateProductRequest struct {
	Name  string  `json:"name" binding:"required"`
	SKU   string  `json:"sku" binding:"required"`
	Price float64 `json:"price" binding:"required,gt=0"`
}

type CreateTransactionRequest struct {
	ProductID int    `json:"product_id" binding:"required"`
	Type      string `json:"type" binding:"required,oneof=IN OUT"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
	Note      string `json:"note"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
