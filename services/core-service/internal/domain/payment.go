package domain

import "time"

// PaymentStatus نوع وضعیت پرداخت
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "PENDING"
	PaymentCompleted PaymentStatus = "COMPLETED"
	PaymentFailed    PaymentStatus = "FAILED"
	PaymentCancelled PaymentStatus = "CANCELLED"
	PaymentVerifying PaymentStatus = "VERIFYING"
)

// Payment نمایش یک تراکنش پرداختی
type Payment struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	OrderID         uint          `gorm:"index;not null" json:"order_id"`
	Order           Order         `gorm:"foreignKey:OrderID" json:"-"`
	UserID          uint          `gorm:"index;not null" json:"user_id"`
	User            User          `gorm:"foreignKey:UserID" json:"-"`
	Amount          float64       `gorm:"type:decimal(15,0);not null" json:"amount"`
	Status          PaymentStatus `gorm:"size:20;default:'PENDING';index" json:"status"`
	PaymentMethod   string        `gorm:"size:50" json:"payment_method"` // card, wallet, etc.
	TransactionID   string        `gorm:"size:100;index" json:"transaction_id"`
	ReferenceNumber string        `gorm:"size:100" json:"reference_number"`
	VerificationURL string        `gorm:"type:text" json:"verification_url"`
	ErrorMessage    string        `gorm:"type:text" json:"error_message"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	VerifiedAt      *time.Time    `json:"verified_at"`
}

// PaymentRepository قرارداد کار با پرداخت‌ها
type PaymentRepository interface {
	Create(ctx interface{}, payment *Payment) error
	GetByID(ctx interface{}, id uint) (*Payment, error)
	GetByOrderID(ctx interface{}, orderID uint) (*Payment, error)
	UpdateStatus(ctx interface{}, id uint, status PaymentStatus) error
	UpdateVerification(ctx interface{}, id uint, status PaymentStatus, verificationURL string) error
	GetByTransactionID(ctx interface{}, transactionID string) (*Payment, error)
}
