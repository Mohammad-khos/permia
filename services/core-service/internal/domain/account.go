package domain

import "time"

// AccountInventory اکانت‌های خریداری شده (خام یا تیم)
type AccountInventory struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ProductSKU   string    `gorm:"size:50;index;not null" json:"product_sku"`
	Email        string    `gorm:"size:200" json:"email"`
	Password     string    `gorm:"size:500" json:"password"`    // رمزنگاری شده
	Additional   string    `gorm:"type:text" json:"additional"` // JSON (Token, Invite Link)
	
	MaxUsers     int       `gorm:"default:1" json:"max_users"`     // ظرفیت کل
	CurrentUsers int       `gorm:"default:0" json:"current_users"` // تعداد فروخته شده
	
	Status       string    `gorm:"size:20;default:'AVAILABLE';index" json:"status"` // AVAILABLE, FILLED, EXPIRED
	PurchasedAt  time.Time `json:"purchased_at"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}