package mocks

import (
	"context"
	"Permia/core-service/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository یک کپی الکی از ریپازیتوری کاربر است
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	args := m.Called(ctx, telegramID)
	// اگر آرگومان اول nil بود، یعنی یوزر پیدا نشد
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) UpdateWallet(ctx context.Context, userID uint, amount float64) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}

// ----------------------------------------------------------------
// متدهای جدید اضافه شده برای هماهنگی با اینترفیس
// ----------------------------------------------------------------

func (m *MockUserRepository) IncrementTotalSpent(ctx context.Context, userID uint, amount float64) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}

func (m *MockUserRepository) GetByReferralCode(ctx context.Context, code string) (*domain.User, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) IncrementReferrals(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}