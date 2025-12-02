package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Permia/core-service/internal/domain"
	"Permia/core-service/internal/handler"
	"Permia/core-service/internal/infrastructure/zarinpal"
	"Permia/core-service/internal/service"
	"Permia/core-service/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestChargePaymentAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// آماده‌سازی داده‌های فیک
	mockOrder := &domain.Order{
		ID:          1,
		OrderNumber: "ORD-001",
		UserID:      1,
		Amount:      100000,
		Status:      domain.OrderPending,
	}

	mockUser := &domain.User{
		ID:            1,
		TelegramID:    123456,
		WalletBalance: 500000,
	}

	// ایجاد Mock Repositories
	mockOrderRepo := new(mocks.MockOrderRepository)
	mockPaymentRepo := new(mocks.MockPaymentRepository)
	mockUserRepo := new(mocks.MockUserRepository)

	mockOrderRepo.On("GetByID", mock.Anything, uint(1)).Return(mockOrder, nil)
	mockPaymentRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *domain.Payment) bool {
		return p.OrderID == 1 && p.UserID == 1 && p.Amount == 100000
	})).Return(nil)
	mockPaymentRepo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockUserRepo.On("GetByID", mock.Anything, uint(1)).Return(mockUser, nil)

	// اپ لاگر را برای تست آماده کن
	testLogger, _ := zap.NewProduction()
	defer testLogger.Sync()

	// ایجاد Zarinpal client
	zarinpalClient := zarinpal.NewZarinpalClient(testLogger)

	// ایجاد سرویس و هندلر
	paymentService := service.NewPaymentService(mockPaymentRepo, mockOrderRepo, mockUserRepo, zarinpalClient, nil, testLogger)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// ایجاد روتر و تعریف مسیر
	router := gin.New()
	router.POST("/api/v1/payment/charge", paymentHandler.ChargePayment)

	// تست ۱: درخواست موفق با روش پرداخت کارت
	t.Run("should charge payment successfully with card", func(t *testing.T) {
		body := map[string]interface{}{
			"order_id":       1,
			"user_id":        1,
			"payment_method": "card",
		}

		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/payment/charge", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["data"])
	})

	// تست ۲: درخواست نامعتبر
	t.Run("should return 400 for invalid request", func(t *testing.T) {
		body := map[string]interface{}{
			"order_id": 1,
			// user_id و payment_method را حذف کردیم
		}

		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/payment/charge", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestVerifyPaymentAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPayment := &domain.Payment{
		ID:            1,
		OrderID:       1,
		UserID:        1,
		Amount:        100000,
		Status:        domain.PaymentVerifying,
		TransactionID: "TXN-1-1234567890",
	}

	mockOrder := &domain.Order{
		ID:     1,
		UserID: 1,
		Amount: 100000,
		Status: domain.OrderPending,
	}

	// ایجاد Mock Repositories
	mockPaymentRepo := new(mocks.MockPaymentRepository)
	mockOrderRepo := new(mocks.MockOrderRepository)
	mockUserRepo := new(mocks.MockUserRepository)

	mockPaymentRepo.On("GetByID", mock.Anything, uint(1)).Return(mockPayment, nil)
	mockPaymentRepo.On("UpdateVerification", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockOrderRepo.On("GetByID", mock.Anything, uint(1)).Return(mockOrder, nil)
	mockOrderRepo.On("UpdateStatus", mock.Anything, uint(1), "PAID").Return(nil)

	// اپ لاگر را برای تست آماده کن
	testLogger, _ := zap.NewProduction()
	defer testLogger.Sync()

	// ایجاد Zarinpal client
	zarinpalClient := zarinpal.NewZarinpalClient(testLogger)

	// ایجاد سرویس و هندلر
	paymentService := service.NewPaymentService(mockPaymentRepo, mockOrderRepo, mockUserRepo, zarinpalClient, nil, testLogger)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// ایجاد روتر
	router := gin.New()
	router.GET("/api/v1/payment/verify", paymentHandler.VerifyPayment)

	// تست ۱: تأیید موفق
	t.Run("should verify payment successfully", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/payment/verify?payment_id=1&ref_id=REF123456", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["data"])
	})

	// تست ۲: پارامتر از دست رفته
	t.Run("should return 400 when ref_id is missing", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/payment/verify?payment_id=1", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// تست ۳: payment_id نامعتبر
	t.Run("should return 400 for invalid payment_id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/payment/verify?payment_id=invalid&ref_id=REF123", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPaymentWalletCharge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockOrder := &domain.Order{
		ID:     1,
		UserID: 1,
		Amount: 50000,
		Status: domain.OrderPending,
	}

	mockUser := &domain.User{
		ID:            1,
		WalletBalance: 100000, // موجودی کافی
	}

	// ایجاد Mocks
	mockOrderRepo := new(mocks.MockOrderRepository)
	mockPaymentRepo := new(mocks.MockPaymentRepository)
	mockUserRepo := new(mocks.MockUserRepository)

	mockOrderRepo.On("GetByID", mock.Anything, uint(1)).Return(mockOrder, nil)
	mockUserRepo.On("GetByID", mock.Anything, uint(1)).Return(mockUser, nil)
	mockUserRepo.On("UpdateWallet", mock.Anything, uint(1), -50000.0).Return(nil)
	mockPaymentRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockPaymentRepo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockOrderRepo.On("UpdateStatus", mock.Anything, uint(1), "PAID").Return(nil)

	// اپ لاگر را برای تست آماده کن
	testLogger, _ := zap.NewProduction()
	defer testLogger.Sync()

	// ایجاد Zarinpal client
	zarinpalClient := zarinpal.NewZarinpalClient(testLogger)

	paymentService := service.NewPaymentService(mockPaymentRepo, mockOrderRepo, mockUserRepo, zarinpalClient, nil, testLogger)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	router := gin.New()
	router.POST("/api/v1/payment/charge", paymentHandler.ChargePayment)

	t.Run("should charge wallet payment successfully", func(t *testing.T) {
		body := map[string]interface{}{
			"order_id":       1,
			"user_id":        1,
			"payment_method": "wallet",
		}

		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/payment/charge", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
