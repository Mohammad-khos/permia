package mocks

import (
	"Permia/core-service/internal/domain"
	"context"

	"github.com/stretchr/testify/mock"
)

// MockOrderRepository یک کپی الکی از ریپازیتوری سفارش است
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, orderID uint, status string) error {
	args := m.Called(ctx, orderID, status)
	return args.Error(0)
}

func (m *MockOrderRepository) GetHistoryByUserID(ctx context.Context, userID uint) ([]domain.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id uint) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) GetAllOrders(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Order), args.Error(1)
}
