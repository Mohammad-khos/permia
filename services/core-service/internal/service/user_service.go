package service

import (
	"Permia/core-service/internal/domain"
	"context"
	"fmt"
	"math/rand"
	"time"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetOrCreateUser کاربر را پیدا می‌کند یا اگر نبود می‌سازد
// اصلاح: پارامتر referralCode به ورودی‌ها اضافه شد
func (s *UserService) GetOrCreateUser(ctx context.Context, telegramID int64, username, firstName, lastName, referralCode string) (*domain.User, error) {
	fmt.Printf("🔍 Checking user: %d\n", telegramID)

	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		fmt.Printf("❌ DB Error during search: %v\n", err)
		return nil, err
	}

	if user != nil {
		fmt.Printf("✅ User Found: %d\n", user.ID)
		// اگر نیاز باشد اطلاعات کاربر (مثل نام کاربری) آپدیت شود، اینجا می‌توان انجام داد
		return user, nil
	}

	fmt.Printf("⚠️ User Not Found. Creating new user...\n")

	// ساخت آبجکت کاربر جدید
	newUser := &domain.User{
		TelegramID:    telegramID,
		Username:      username,
		FirstName:     firstName,
		LastName:      lastName,
		WalletBalance: 0,
		// اصلاح: تولید کد دعوت رندوم و کوتاه (بهتر از ترکیب یوزرنیم است)
		ReferralCode: s.generateReferralCode(8),
		CreatedAt:    time.Now(),
	}

	// ۴. بررسی کد معرف (اگر ارسال شده باشد)
	if referralCode != "" {
		// اصلاح: استفاده از s.repo به جای s.userRepo
		referrer, err := s.repo.GetByReferralCode(ctx, referralCode)
		
		// شرط: معرف پیدا شود و کاربر خودش را دعوت نکرده باشد
		if err == nil && referrer != nil && referrer.TelegramID != telegramID {
			newUser.ReferredBy = &referrer.ID

			// افزایش آمار معرف
			if err := s.repo.IncrementReferrals(ctx, referrer.ID); err != nil {
				fmt.Printf("⚠️ Failed to increment referrals: %v\n", err)
			}
		}
	}

	if err := s.repo.Create(ctx, newUser); err != nil {
		fmt.Printf("❌ Create Error: %v\n", err)
		return nil, err
	}

	fmt.Printf("🎉 User Created Successfully: %d\n", newUser.ID)
	return newUser, nil
}

// GetByTelegramID فقط اطلاعات کاربر را می‌گیرد
func (s *UserService) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	return s.repo.GetByTelegramID(ctx, telegramID)
}

func (s *UserService) GetBalance(ctx context.Context, userID uint) (float64, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return user.WalletBalance, nil
}

func (s *UserService) GetByID(ctx context.Context, userID uint) (*domain.User, error) {
	return s.repo.GetByID(ctx, userID)
}

// تابع کمکی برای تولید کد دعوت رندوم
func (s *UserService) generateReferralCode(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	// مقداردهی اولیه سید رندوم (برای هر بار اجرا متفاوت باشد)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}