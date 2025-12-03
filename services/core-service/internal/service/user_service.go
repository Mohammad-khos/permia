package service

import (
	"Permia/core-service/internal/domain"
	"context"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetOrCreateUser کاربر را پیدا می‌کند یا اگر نبود می‌سازد
func (s *UserService) GetOrCreateUser(ctx context.Context, telegramID int64, username, firstName, lastName string) (*domain.User, error) {
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err == nil {
		return user, nil
	}

	newUser := &domain.User{
		TelegramID: telegramID,
		Username:   username,
		FirstName:  firstName,
		LastName:   lastName,
		WalletBalance:    0,
	}

	if err := s.repo.Create(ctx, newUser); err != nil {
		return nil, err
	}

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