package mocks

import (
	"Permia/core-service/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockPaymentRepository یک کپی الکی از ریپازیتوری پرداخت است
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(ctx interface{}, payment *domain.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByID(ctx interface{}, id uint) (*domain.Payment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetByOrderID(ctx interface{}, orderID uint) (*domain.Payment, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) UpdateStatus(ctx interface{}, id uint, status domain.PaymentStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockPaymentRepository) UpdateVerification(ctx interface{}, id uint, status domain.PaymentStatus, verificationURL string) error {
	args := m.Called(ctx, id, status, verificationURL)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByTransactionID(ctx interface{}, transactionID string) (*domain.Payment, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}
