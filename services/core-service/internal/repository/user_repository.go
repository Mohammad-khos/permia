package repository

import (
	"context"
	"errors"
	"Permia/core-service/internal/domain"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository سازنده ریپازیتوری کاربر
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// Create ایجاد کاربر جدید
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByTelegramID پیدا کردن کاربر با آیدی تلگرام
func (r *userRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("telegram_id = ?", telegramID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // کاربر نیست، ارور هم نده
		}
		return nil, err
	}
	return &user, nil
}

// GetByID پیدا کردن با ID دیتابیس
func (r *userRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateWallet آپدیت موجودی کیف پول (با تراکنش امن)
func (r *userRepository) UpdateWallet(ctx context.Context, userID uint, amount float64) error {
	// استفاده از gorm.Expr برای جلوگیری از Race Condition
	// wallet_balance = wallet_balance + amount
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Update("wallet_balance", gorm.Expr("wallet_balance + ?", amount)).Error
}

// به انتهای فایل اضافه کنید

// IncrementTotalSpent مجموع خرید کاربر را افزایش می‌دهد
func (r *userRepository) IncrementTotalSpent(ctx context.Context, userID uint, amount float64) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Update("total_spent", gorm.Expr("total_spent + ?", amount)).Error
}

// GetByReferralCode پیدا کردن کاربر با کد دعوت
func (r *userRepository) GetByReferralCode(ctx context.Context, code string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("referral_code = ?", code).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// IncrementReferrals افزایش تعداد دعوت‌شده‌ها
func (r *userRepository) IncrementReferrals(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Update("total_referrals", gorm.Expr("total_referrals + 1")).Error
}