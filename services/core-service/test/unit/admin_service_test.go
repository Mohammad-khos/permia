package unit

import (
	"context"
	"errors"
	"testing"

	"Permia/core-service/internal/domain"
	"Permia/core-service/internal/service"
	"Permia/core-service/test/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestAdminServiceAddInventory(t *testing.T) {
	// آماده‌سازی Mocks
	mockAccountRepo := new(mocks.MockAccountRepository)
	mockProductRepo := new(mocks.MockProductRepository)
	mockOrderRepo := new(mocks.MockOrderRepository)

	mockProduct := &domain.Product{
		ID:    1,
		SKU:   "gpt_shared",
		Title: "GPT Shared",
		Price: 100000,
	}

	// تست ۱: اضافه کردن موجودی موفق
	t.Run("should add inventory successfully", func(t *testing.T) {
		mockProductRepo.On("GetBySKU", mock.Anything, "gpt_shared").Return(mockProduct, nil)
		mockAccountRepo.On("CreateBatch", mock.Anything, mock.MatchedBy(func(accounts []domain.AccountInventory) bool {
			return len(accounts) == 5
		})).Return(nil)

		testLogger, _ := zap.NewProduction()
		defer testLogger.Sync()

		adminService := service.NewAdminService(mockAccountRepo, mockOrderRepo, mockProductRepo, nil, testLogger)

		req := &domain.AdminInventoryRequest{
			ProductSKU: "gpt_shared",
			Email:      "account",
			Password:   "pass",
			MaxUsers:   3,
			Count:      5,
		}

		result, err := adminService.AddInventory(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "gpt_shared", result["product_sku"])
		assert.Equal(t, 5, result["added_count"])
	})

	// تست ۲: محصول یافت نشد
	t.Run("should return error when product not found", func(t *testing.T) {
		mockProductRepo2 := new(mocks.MockProductRepository)
		mockProductRepo2.On("GetBySKU", mock.Anything, "invalid").Return(nil, gorm.ErrRecordNotFound)

		testLogger, _ := zap.NewProduction()
		defer testLogger.Sync()

		adminService := service.NewAdminService(mockAccountRepo, mockOrderRepo, mockProductRepo2, nil, testLogger)

		req := &domain.AdminInventoryRequest{
			ProductSKU: "invalid",
			Email:      "account",
			Count:      1,
		}

		result, err := adminService.AddInventory(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "محصول پیدا نشد")
	})

	// تست ۳: خرابی در ایجاد اکانت‌ها
	t.Run("should return error when batch creation fails", func(t *testing.T) {
		mockAccountRepo2 := new(mocks.MockAccountRepository)
		mockProductRepo3 := new(mocks.MockProductRepository)

		mockProductRepo3.On("GetBySKU", mock.Anything, "gpt_shared").Return(mockProduct, nil)
		mockAccountRepo2.On("CreateBatch", mock.Anything, mock.Anything).Return(errors.New("database error"))

		testLogger, _ := zap.NewProduction()
		defer testLogger.Sync()

		adminService := service.NewAdminService(mockAccountRepo2, mockOrderRepo, mockProductRepo3, nil, testLogger)

		req := &domain.AdminInventoryRequest{
			ProductSKU: "gpt_shared",
			Email:      "account",
			Count:      1,
		}

		result, err := adminService.AddInventory(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestAdminServiceCompleteOrder(t *testing.T) {
	mockAccountRepo := new(mocks.MockAccountRepository)
	mockProductRepo := new(mocks.MockProductRepository)

	mockOrder := &domain.Order{
		ID:     1,
		UserID: 1,
		Amount: 100000,
		Status: domain.OrderPaid,
	}

	// تست ۱: تکمیل سفارش موفق
	t.Run("should complete order successfully", func(t *testing.T) {
		mockOrderRepo2 := new(mocks.MockOrderRepository)
		mockOrderRepo2.On("GetByID", mock.Anything, uint(1)).Return(mockOrder, nil)
		mockOrderRepo2.On("UpdateStatus", mock.Anything, uint(1), "COMPLETED").Return(nil)

		testLogger, _ := zap.NewProduction()
		defer testLogger.Sync()

		adminService := service.NewAdminService(mockAccountRepo, mockOrderRepo2, mockProductRepo, nil, testLogger)

		result, err := adminService.CompleteOrder(context.Background(), 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result["order_id"])
	})

	// تست ۲: سفارش یافت نشد
	t.Run("should return error when order not found", func(t *testing.T) {
		mockOrderRepo3 := new(mocks.MockOrderRepository)
		mockOrderRepo3.On("GetByID", mock.Anything, uint(999)).Return(nil, gorm.ErrRecordNotFound)

		testLogger, _ := zap.NewProduction()
		defer testLogger.Sync()

		adminService := service.NewAdminService(mockAccountRepo, mockOrderRepo3, mockProductRepo, nil, testLogger)

		result, err := adminService.CompleteOrder(context.Background(), 999)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestAdminServiceGetOrderStats(t *testing.T) {
	// تست فقط ساختار است؛ برای دریافت آمار، DB بیشتر ضروری است
	t.Run("should create stats structure", func(t *testing.T) {
		testLogger, _ := zap.NewProduction()
		defer testLogger.Sync()

		adminService := service.NewAdminService(nil, nil, nil, nil, testLogger)
		assert.NotNil(t, adminService)
	})
}
