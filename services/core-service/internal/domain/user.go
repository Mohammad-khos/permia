package domain

import "time"

type User struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TelegramID     int64     `gorm:"uniqueIndex;not null" json:"telegram_id"`
	Username       string    `gorm:"size:100" json:"username"`
	FirstName      string    `gorm:"size:100" json:"first_name"`
	LastName       string    `gorm:"size:100" json:"last_name"`
	WalletBalance  float64   `gorm:"type:decimal(15,0);default:0" json:"wallet_balance"`
	TotalSpent     float64   `gorm:"type:decimal(15,0);default:0" json:"total_spent"`
	ReferralCode   string    `gorm:"size:20;uniqueIndex" json:"referral_code"`
	ReferredBy     *uint     `json:"referred_by"` // ID معرف
	TotalReferrals int       `gorm:"default:0" json:"total_referrals"`
	IsBanned       bool      `gorm:"default:false" json:"is_banned"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}