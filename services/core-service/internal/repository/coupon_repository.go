package repository

import (
	"Permia/core-service/internal/domain"
	"context"
	"time"

	"gorm.io/gorm"
)

type couponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) domain.CouponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) GetByCode(ctx context.Context, code string) (*domain.Coupon, error) {
	var coupon domain.Coupon
	// کوپن باید پیدا شود، منقضی نشده باشد و محدودیت استفاده نداشته باشد
	err := r.db.WithContext(ctx).
		Where("code = ?", code).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		First(&coupon).Error
	return &coupon, err
}

func (r *couponRepository) GetByUserID(ctx context.Context, userID uint) ([]domain.Coupon, error) {
	var coupons []domain.Coupon
	// کوپن‌هایی که یا عمومی هستند یا اختصاصی برای این کاربر
	// و هنوز منقضی نشده‌اند
	err := r.db.WithContext(ctx).
		Where("(assignee_id = ? OR assignee_id IS NULL)", userID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Where("used_count < usage_limit").
		Find(&coupons).Error
	return coupons, err
}

func (r *couponRepository) IncrementUsage(ctx context.Context, couponID uint) error {
	return r.db.WithContext(ctx).Model(&domain.Coupon{}).
		Where("id = ?", couponID).
		Update("used_count", gorm.Expr("used_count + 1")).Error
}