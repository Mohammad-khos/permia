package domain

import (
	"context"
)

// UserRepository قرارداد کار با دیتابیس کاربر
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	UpdateWallet(ctx context.Context, userID uint, amount float64) error
	GetByID(ctx context.Context, id uint) (*User, error)
	IncrementTotalSpent(ctx context.Context, userID uint, amount float64) error
}

// ProductRepository قرارداد کار با محصولات
type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	GetActiveProducts(ctx context.Context) ([]Product, error)
	GetBySKU(ctx context.Context, sku string) (*Product, error)
}

// OrderRepository قرارداد کار با سفارشات
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	UpdateStatus(ctx context.Context, orderID uint, status string) error
	GetHistoryByUserID(ctx context.Context, userID uint) ([]Order, error)
	GetByID(ctx context.Context, id uint) (*Order, error)
	GetAllOrders(ctx context.Context) ([]Order, error)
}

// AccountRepository قرارداد کار با انبار اکانت‌ها
type AccountRepository interface {
	GetAvailableAccount(ctx context.Context, productSKU string) (*AccountInventory, error)
	MarkAsSold(ctx context.Context, accountID uint) error
	CreateBatch(ctx context.Context, accounts []AccountInventory) error
}

type VirtualCard struct {
	ID     string
	PAN    string // شماره کارت
	CVV    string
	Expiry string
}

// VCCProvider قرارداد اتصال به سرویس‌دهنده کارت (Brocard)
type VCCProvider interface {
	// IssueCard کارت صادر می‌کند (BIN در بروکارد معمولا با ID نوع کارت مشخص می‌شود)
	IssueCard(amount float64, cardTypeID string) (*VirtualCard, error)
	GetBalance() (float64, error)
}
