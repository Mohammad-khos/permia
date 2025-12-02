package mocks

import (
	"context"
	"Permia/core-service/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockProductRepository شبیه‌ساز دیتابیس محصولات
type MockProductRepository struct {
	mock.Mock
}

// GetActiveProducts شبیه‌سازی دریافت لیست محصولات
func (m *MockProductRepository) GetActiveProducts(ctx context.Context) ([]domain.Product, error) {
	args := m.Called(ctx)
	// تبدیل آرگومان اول به اسلایس محصولات
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Product), args.Error(1)
}

// GetBySKU شبیه‌سازی دریافت یک محصول
func (m *MockProductRepository) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

// Create شبیه‌سازی ساخت محصول (فقط برای اینکه اینترفیس کامل شود)
func (m *MockProductRepository) Create(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}