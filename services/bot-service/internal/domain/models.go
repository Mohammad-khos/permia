package domain

import "time"

// User represents a user in the system.
type User struct {
	ID             uint      `json:"id"`
	TelegramID     int64     `json:"telegram_id"`
	Username       string    `json:"username"`
	Balance        float64   `json:"balance"`
	ReferralCode   string    `json:"referral_code"`
	TotalReferrals int       `json:"total_referrals"`
	CreatedAt      time.Time `json:"created_at"`
}

// Product represents a sellable item.
type Product struct {
	ID          uint    `json:"id"`
	SKU         string  `json:"sku"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Type        string  `json:"type"`
}

// Coupon represents a discount code.
type Coupon struct {
	Code        string  `json:"code"`
	Percent     float64 `json:"percent"`
	MaxDiscount float64 `json:"max_discount"`
}

// Subscription represents a user's active subscription.
type Subscription struct {
	ID            uint    `json:"id"`
	ProductName   string  `json:"ProductName"`
	Sku           string  `json:"sku"`
	DeliveredData string  `json:"DeliveredData"`
	Amount        float64 `json:"amount"`
	CreatedAt     string  `json:"CreatedAt"`
	ExpiresAt     string  `json:"ExpiresAt"`
}

// Order represents a user's purchase.
type Order struct {
	ID            uint      `json:"id"`
	OrderID       uint      `json:"order_id"`
	UserID        uint      `json:"user_id"`
	ProductSKU    string    `json:"product_sku"`
	Amount        float64   `json:"amount"`
	DeliveredData string    `json:"delivered_data"`
	CreatedAt     time.Time `json:"created_at"`
}

// UserState represents the state of a user in a conversation.
type UserState int

const (
	StateNone             UserState = iota
	StateWaitingForAmount           // شارژ کیف پول
	StateWaitingForCoupon           // منتظر کد تخفیف
)