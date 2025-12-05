package service

import (
	"Permia/core-service/internal/domain"
	"context"
	"errors"

)

type CouponService struct {
	repo domain.CouponRepository
}

func NewCouponService(repo domain.CouponRepository) *CouponService {
	return &CouponService{repo: repo}
}

// ValidateCoupon بررسی اعتبار و محاسبه قیمت جدید
func (s *CouponService) ValidateCoupon(ctx context.Context, code string, userID uint, originalPrice float64) (float64, float64, error) {
	coupon, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return 0, 0, errors.New("کد تخفیف نامعتبر است")
	}

	// ۱. بررسی مالکیت (اگر اختصاصی باشد)
	if coupon.AssigneeID != nil && *coupon.AssigneeID != userID {
		return 0, 0, errors.New("این کد تخفیف متعلق به شما نیست")
	}

	// ۲. بررسی محدودیت تعداد
	if coupon.UsageLimit > 0 && coupon.UsedCount >= coupon.UsageLimit {
		return 0, 0, errors.New("مهلت استفاده از این کد تمام شده است")
	}

	// ۳. محاسبه تخفیف
	discountAmount := (originalPrice * coupon.Percent) / 100
	if coupon.MaxDiscount > 0 && discountAmount > coupon.MaxDiscount {
		discountAmount = coupon.MaxDiscount
	}

	finalPrice := originalPrice - discountAmount
	if finalPrice < 0 {
		finalPrice = 0
	}

	return finalPrice, discountAmount, nil
}

func (s *CouponService) GetUserCoupons(ctx context.Context, userID uint) ([]domain.Coupon, error) {
	return s.repo.GetByUserID(ctx, userID)
}