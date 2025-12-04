package domain

import "time"

// User represents a user in the system.
type User struct {
	ID         uint      `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username"`
	Balance    float64   `json:"wallet_balance"`
	CreatedAt  time.Time `json:"created_at"`
}

// Product represents a sellable item.
type Product struct {
	ID          uint    `json:"id"`
	SKU         string  `json:"sku"`
	Title        string  `json:"title"`
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

// Subscription ساختار اطلاعات اشتراک دریافتی از Core Service
type Subscription struct {
	ID            int64   `json:"id"`
	ProductName   string  `json:"ProductName"`   // نام محصول
	Sku           string  `json:"sku"`           // کد محصول
	DeliveredData string  `json:"DeliveredData"` // اطلاعات تحویل (یوزر/پسورد)
	Amount        float64 `json:"amount"`
	CreatedAt     string  `json:"CreatedAt"`     // تاریخ ایجاد (رشته)
	ExpiresAt     string  `json:"ExpiresAt"`     // تاریخ انقضا (رشته)
}

// UserState represents the state of a user in a conversation.
type UserState int

const (
	StateNone UserState = iota
	StateWaitingForAmount
)