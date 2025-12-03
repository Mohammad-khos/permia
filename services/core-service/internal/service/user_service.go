package service

import (
	"Permia/core-service/internal/domain"
	"context"
	// "errors"
	"fmt"
	// "strings"

	// "gorm.io/gorm"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetOrCreateUser کاربر را پیدا می‌کند یا اگر نبود می‌سازد
func (s *UserService) GetOrCreateUser(ctx context.Context, telegramID int64, username, firstName, lastName string) (*domain.User, error) {
	fmt.Printf("🔍 Checking user: %d\n", telegramID) // لاگ 1

	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		fmt.Printf("❌ DB Error during search: %v\n", err)
		return nil, err
	}
	
	if user != nil {
		fmt.Printf("✅ User Found: %d\n", user.ID)
		return user, nil
	}

	fmt.Printf("⚠️ User Not Found. Creating new user...\n") // لاگ 2

	newUser := &domain.User{
		TelegramID:    telegramID,
		Username:      username,
		FirstName:     firstName,
		LastName:      lastName,
		WalletBalance: 0,
		ReferralCode:  fmt.Sprintf("%s_%d", username, telegramID), // کد ریفرال ساده
	}

	if err := s.repo.Create(ctx, newUser); err != nil {
		fmt.Printf("❌ Create Error: %v\n", err)
		return nil, err
	}

	fmt.Printf("🎉 User Created Successfully: %d\n", newUser.ID) // لاگ 3
	return newUser, nil
}

// GetByTelegramID فقط اطلاعات کاربر را می‌گیرد (متد جدید)
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