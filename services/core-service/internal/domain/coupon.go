package domain

import (
	"context"
	"time"
)

type Coupon struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Code           string     `gorm:"uniqueIndex;size:50;not null" json:"code"` // کد تخفیف (مثلا OFF50)
	Percent        float64    `json:"percent"`                                  // درصد تخفیف (۰ تا ۱۰۰)
	MaxDiscount    float64    `json:"max_discount"`                             // سقف تخفیف (تومان)
	UsageLimit     int        `json:"usage_limit"`                              // محدودیت تعداد استفاده کلی
	UsedCount      int        `json:"used_count"`                               // تعداد استفاده شده تا الان
	ExpiresAt      *time.Time `json:"expires_at"`                               // تاریخ انقضا
	AssigneeID     *uint      `json:"assignee_id"`                              // اگر پر باشد، فقط برای این کاربر خاص است
	CreatedAt      time.Time  `json:"created_at"`
}

// اینترفیس ریپازیتوری
type CouponRepository interface {
	GetByCode(ctx context.Context, code string) (*Coupon, error)
	GetByUserID(ctx context.Context, userID uint) ([]Coupon, error)
	IncrementUsage(ctx context.Context, couponID uint) error
}