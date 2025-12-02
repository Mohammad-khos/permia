package domain

import "time"

type Product struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	SKU          string  `gorm:"uniqueIndex;size:50;not null" json:"sku"` // شناسه فنی (gpt_shared_4)
	Category     string  `gorm:"size:50;index" json:"category"`           // chatgpt, gemini, claude, tools
	Title        string  `gorm:"size:200;not null" json:"title"`
	Description  string  `gorm:"type:text" json:"description"`
	Price        float64 `gorm:"type:decimal(15,0);not null" json:"price"`
	Type         string  `gorm:"size:50;not null" json:"type"`      // shared, private_legal, private_invite, ready_made
	Capacity     int     `gorm:"default:1" json:"capacity"`         // ظرفیت صندلی (1, 3, 4)
	IsActive     bool    `gorm:"default:true" json:"is_active"`     // برای ناموجود کردن
	DisplayOrder int     `gorm:"default:0" json:"display_order"`    // ترتیب نمایش
	CreatedAt    time.Time
	UpdatedAt    time.Time
}