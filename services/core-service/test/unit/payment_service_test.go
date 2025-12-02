package unit

import (
	"context"
	"testing"

	"Permia/core-service/internal/domain"
	"Permia/core-service/internal/infrastructure/zarinpal"
	"Permia/core-service/internal/service"
	"Permia/core-service/test/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestPaymentServiceCharge(t *testing.T) {
	mockOrder := &domain.Order{
		ID:     1,
		UserID: 1,
		Amount: 100000,
		Status: domain.OrderPending,
	}

	mockUser := &domain.User{
		ID:            1,
		WalletBalance: 500000,
	}

	// تست ۱: پرداخت موفق با کارت
	t.Run("should charge payment successfully with card", func(t *testing.T) {
		mockOrderRepo2 := new(mocks.MockOrderRepository)
		mockUserRepo2 := new(mocks.MockUserRepository)
		mockPaymentRepo2 := new(mocks.MockPaymentRepository)

		mockOrderRepo2.On("GetByID", mock.Anything, uint(1)).Return(mockOrder, nil)
		mockUserRepo2.On("GetByID", mock.Anything, uint(1)).Return(mockUser, nil)
		mockPaymentRepo2.On("Create", mock.Anything, mock.Anything).Return(nil)
		mockPaymentRepo2.On("UpdateStatus", mock.Anything, mock.Anything, domain.PaymentVerifying).Return(nil)

		testLogger, _ := zap.NewProduction()
		defer testLogger.Sync()
		zarinpalClient := zarinpal.NewZarinpalClient(testLogger)

		paymentService := service.NewPaymentService(mockPaymentRepo2, mockOrderRepo2, mockUserRepo2, zarinpalClient, nil, testLogger)

		req := &service.ChargeRequest{
			OrderID:       1,
			UserID:        1,
			PaymentMethod: "card",
		}

		result, err := paymentService.Charge(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.OrderID)
		assert.Equal(t, string(domain.PaymentVerifying), result.Status)
	})

	// تست ۲: سفارش یافت نشد
	t.Run("should return error when order not found", func(t *testing.T) {
		mockOrderRepo3 := new(mocks.MockOrderRepository)
		mockPaymentRepo3 := new(mocks.MockPaymentRepository)
		mockOrderRepo3.On("GetByID", mock.Anything, uint(999)).Return(nil, gorm.ErrRecordNotFound)

		testLogger, _ := zap.NewProduction()
		defer testLogger.Sync()
		zarinpalClient := zarinpal.NewZarinpalClient(testLogger)

		paymentService := service.NewPaymentService(mockPaymentRepo3, mockOrderRepo3, nil, zarinpalClient, nil, testLogger)

		req := &service.ChargeRequest{
			OrderID:       999,
			UserID:        1,
			PaymentMethod: "card",
		}

		result, err := paymentService.Charge(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	// تست ۳: موجودی کیف پول کافی نیست
	t.Run("should return error when wallet balance is insufficient", func(t *testing.T) {
		mockOrderRepo4 := new(mocks.MockOrderRepository)
		mockUserRepo4 := new(mocks.MockUserRepository)
		mockPaymentRepo4 := new(mocks.MockPaymentRepository)

		mockOrderRepo4.On("GetByID", mock.Anything, uint(1)).Return(mockOrder, nil)

		insufficientUser := &domain.User{
			ID:            1,
			WalletBalance: 50000, // کمتر از Amount
		}
		mockUserRepo4.On("GetByID", mock.Anything, uint(1)).Return(insufficientUser, nil)

		testLogger, _ := zap.NewProduction()
		defer testLogger.Sync()
		zarinpalClient := zarinpal.NewZarinpalClient(testLogger)

		paymentService := service.NewPaymentService(mockPaymentRepo4, mockOrderRepo4, mockUserRepo4, zarinpalClient, nil, testLogger)

		req := &service.ChargeRequest{
			OrderID:       1,
			UserID:        1,
			PaymentMethod: "wallet",
		}

		result, err := paymentService.Charge(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "موجودی")
	})
}

func TestPaymentServiceVerify(t *testing.T) {
	mockPayment := &domain.Payment{
		ID:            1,
		OrderID:       1,
		UserID:        1,
		Amount:        100000,
		Status:        domain.PaymentVerifying,
		TransactionID: "TXN-1-123",
	}

	mockOrder := &domain.Order{
		ID:     1,
		UserID: 1,
		Amount: 100000,
		Status: domain.OrderPending,
	}

	// تست ۱: تأیید موفق
	t.Run("should verify payment successfully", func(t *testing.T) {
		mockPaymentRepo2 := new(mocks.MockPaymentRepository)
		mockOrderRepo2 := new(mocks.MockOrderRepository)

		mockPaymentRepo2.On("GetByID", mock.Anything, uint(1)).Return(mockPayment, nil)
		mockPaymentRepo2.On("UpdateVerification", mock.Anything, uint(1), domain.PaymentCompleted, "REF123").Return(nil)
		mockOrderRepo2.On("GetByID", mock.Anything, uint(1)).Return(mockOrder, nil)
		mockOrderRepo2.On("UpdateStatus", mock.Anything, uint(1), "PAID").Return(nil)

		testLogger, _ := zap.NewProduction()
		defer testLogger.Sync()
		zarinpalClient := zarinpal.NewZarinpalClient(testLogger)

		paymentService := service.NewPaymentService(mockPaymentRepo2, mockOrderRepo2, nil, zarinpalClient, nil, testLogger)

		req := &service.VerifyRequest{
			PaymentID: 1,
			Authority: "REF123",
		}

		result, err := paymentService.Verify(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.PaymentID)
		assert.Equal(t, string(domain.PaymentCompleted), result.Status)
	})

	// تست ۲: پرداخت یافت نشد
	t.Run("should return error when payment not found", func(t *testing.T) {
		mockPaymentRepo3 := new(mocks.MockPaymentRepository)
		mockPaymentRepo3.On("GetByID", mock.Anything, uint(999)).Return(nil, gorm.ErrRecordNotFound)

		testLogger, _ := zap.NewProduction()
		defer testLogger.Sync()
		zarinpalClient := zarinpal.NewZarinpalClient(testLogger)

		paymentService := service.NewPaymentService(mockPaymentRepo3, nil, nil, zarinpalClient, nil, testLogger)

		req := &service.VerifyRequest{
			PaymentID: 999,
			Authority: "REF123",
		}

		result, err := paymentService.Verify(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	// تست ۳: پرداخت در وضعیت غلط
	t.Run("should return error when payment is not in verifying state", func(t *testing.T) {
		mockPaymentRepo4 := new(mocks.MockPaymentRepository)

		completedPayment := &domain.Payment{
			ID:     1,
			Status: domain.PaymentCompleted,
		}
		mockPaymentRepo4.On("GetByID", mock.Anything, uint(1)).Return(completedPayment, nil)

		testLogger, _ := zap.NewProduction()
		defer testLogger.Sync()
		zarinpalClient := zarinpal.NewZarinpalClient(testLogger)

		paymentService := service.NewPaymentService(mockPaymentRepo4, nil, nil, zarinpalClient, nil, testLogger)

		req := &service.VerifyRequest{
			PaymentID: 1,
			Authority: "REF123",
		}

		result, err := paymentService.Verify(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestPaymentServiceInvalidRefID(t *testing.T) {
	mockPayment := &domain.Payment{
		ID:      1,
		Status:  domain.PaymentVerifying,
		OrderID: 1,
	}

	t.Run("should handle invalid ref_id", func(t *testing.T) {
		mockPaymentRepo2 := new(mocks.MockPaymentRepository)
		mockPaymentRepo2.On("GetByID", mock.Anything, uint(1)).Return(mockPayment, nil)
		mockPaymentRepo2.On("UpdateStatus", mock.Anything, uint(1), domain.PaymentFailed).Return(nil)

		testLogger, _ := zap.NewProduction()
		defer testLogger.Sync()
		zarinpalClient := zarinpal.NewZarinpalClient(testLogger)

		paymentService := service.NewPaymentService(mockPaymentRepo2, nil, nil, zarinpalClient, nil, testLogger)

		req := &service.VerifyRequest{
			PaymentID: 1,
			Authority: "", // Authority خالی
		}

		result, err := paymentService.Verify(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, string(domain.PaymentFailed), result.Status)
	})
}
