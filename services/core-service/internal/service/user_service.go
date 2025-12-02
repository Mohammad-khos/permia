package service

import (
	"context"
	"Permia/core-service/internal/domain"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetOrCreateUser اگر کاربر هست برش‌گردان، اگر نیست بساز
func (s *UserService) GetOrCreateUser(ctx context.Context, telegramID int64, username, firstName, lastName string) (*domain.User, error) {
	// 1. جستجو
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}
	
	// 2. اگر پیدا شد، برگردان
	if user != nil {
		return user, nil
	}

	// 3. اگر نبود، بساز
	newUser := &domain.User{
		TelegramID: telegramID,
		Username:   username,
		FirstName:  firstName,
		LastName:   lastName,
		// کد ریفرال رو بعدا با پکیج random میسازیم
		ReferralCode: username + "_ref", 
	}
	
	if err := s.repo.Create(ctx, newUser); err != nil {
		return nil, err
	}
	
	return newUser, nil
}

// GetBalance گرفتن موجودی
func (s *UserService) GetBalance(ctx context.Context, userID uint) (float64, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return user.WalletBalance, nil
}