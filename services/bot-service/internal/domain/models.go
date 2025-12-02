package domain

import "time"

// User represents a user in the system.
type User struct {
	ID         uint      `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username"`
	Balance    float64   `json:"balance"`
	CreatedAt  time.Time `json:"created_at"`
}

// Product represents a sellable item.
type Product struct {
	ID          uint    `json:"id"`
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
}

// Order represents a user's purchase.
type Order struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	ProductSKU string    `json:"product_sku"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}