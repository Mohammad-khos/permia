package domain

import "time"

// تعریف نوع اختصاصی برای وضعیت سفارش
type OrderStatus string

// تعریف ثابت‌ها (حتما با حرف بزرگ شروع شوند تا Export شوند)
const (
	OrderPending   OrderStatus = "PENDING"
	OrderPaid      OrderStatus = "PAID"
	OrderCompleted OrderStatus = "COMPLETED"
	OrderFailed    OrderStatus = "FAILED"
	OrderCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	OrderNumber string `gorm:"uniqueIndex;size:50" json:"order_number"`
	UserID      uint   `gorm:"index;not null" json:"user_id"`
	User        User   `gorm:"foreignKey:UserID" json:"-"`

	ProductID uint    `gorm:"index;not null" json:"product_id"`
	Product   Product `gorm:"foreignKey:ProductID" json:"-"`

	AccountID *uint            `gorm:"index" json:"account_id"`
	Account   AccountInventory `gorm:"foreignKey:AccountID" json:"-"`

	Amount float64 `gorm:"type:decimal(15,0);not null" json:"amount"`

	Status OrderStatus `gorm:"size:20;default:'PENDING';index" json:"status"`

	PaymentMethod string `gorm:"size:50" json:"payment_method"`
	DeliveredData string `gorm:"type:text" json:"delivered_data"`

	CreatedAt   time.Time  `json:"created_at"`
	DeliveredAt *time.Time `json:"delivered_at"`
}
