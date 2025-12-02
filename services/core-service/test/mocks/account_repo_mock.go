package mocks

import (
	"Permia/core-service/internal/domain"
	"context"

	"github.com/stretchr/testify/mock"
)

// MockAccountRepository یک کپی الکی از ریپازیتوری انبار اکانت است
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) GetAvailableAccount(ctx context.Context, productSKU string) (*domain.AccountInventory, error) {
	args := m.Called(ctx, productSKU)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AccountInventory), args.Error(1)
}

func (m *MockAccountRepository) MarkAsSold(ctx context.Context, accountID uint) error {
	args := m.Called(ctx, accountID)
	return args.Error(0)
}

func (m *MockAccountRepository) CreateBatch(ctx context.Context, accounts []domain.AccountInventory) error {
	args := m.Called(ctx, accounts)
	return args.Error(0)
}
