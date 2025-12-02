package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"Permia/core-service/internal/domain"
	"Permia/core-service/internal/handler"
	"Permia/core-service/internal/service"
	"Permia/core-service/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestGetOrdersAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// داده‌های فیک
	mockOrders := []domain.Order{
		{
			ID:          1,
			OrderNumber: "ORD-001",
			UserID:      1,
			ProductID:   1,
			Amount:      100000,
			Status:      domain.OrderCompleted,
			CreatedAt:   time.Now(),
			Product: domain.Product{
				ID:    1,
				SKU:   "gpt_shared",
				Title: "GPT Shared",
				Price: 100000,
			},
		},
		{
			ID:          2,
			OrderNumber: "ORD-002",
			UserID:      1,
			ProductID:   2,
			Amount:      50000,
			Status:      domain.OrderPending,
			CreatedAt:   time.Now(),
			Product: domain.Product{
				ID:    2,
				SKU:   "gemini_basic",
				Title: "Gemini Basic",
				Price: 50000,
			},
		},
	}

	// ایجاد Mocks
	mockOrderRepo := new(mocks.MockOrderRepository)
	mockUserRepo := new(mocks.MockUserRepository)

	mockOrderRepo.On("GetAllOrders", mock.Anything).Return(mockOrders, nil)

	// ایجاد سرویس و هندلر
	orderService := service.NewOrderService(mockOrderRepo, mockUserRepo, nil, nil, nil, nil)
	orderHandler := handler.NewOrderHandler(orderService, nil)

	// ایجاد روتر
	router := gin.New()
	router.GET("/api/v1/orders", orderHandler.GetOrders)

	// تست: دریافت لیست سفارشات
	t.Run("should get all orders successfully", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/orders", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["data"])
	})
}

func TestGetOrderByIDAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockOrder := &domain.Order{
		ID:            1,
		OrderNumber:   "ORD-001",
		UserID:        1,
		ProductID:     1,
		Amount:        100000,
		Status:        domain.OrderCompleted,
		DeliveredData: "Email: test@example.com\nPassword: xxx",
		CreatedAt:     time.Now(),
		Product: domain.Product{
			ID:    1,
			SKU:   "gpt_shared",
			Title: "GPT Shared",
			Price: 100000,
		},
	}

	// ایجاد Mocks
	mockOrderRepo := new(mocks.MockOrderRepository)
	mockUserRepo := new(mocks.MockUserRepository)

	mockOrderRepo.On("GetByID", mock.Anything, uint(1)).Return(mockOrder, nil)
	mockOrderRepo.On("GetByID", mock.Anything, uint(999)).Return(nil, gorm.ErrRecordNotFound)

	// ایجاد سرویس و هندلر
	orderService := service.NewOrderService(mockOrderRepo, mockUserRepo, nil, nil, nil, nil)
	orderHandler := handler.NewOrderHandler(orderService, nil)

	// ایجاد روتر
	router := gin.New()
	router.GET("/api/v1/orders/:id", orderHandler.GetOrderByID)

	// تست ۱: دریافت سفارش موجود
	t.Run("should get order by ID successfully", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/orders/1", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["id"])
		assert.Equal(t, "ORD-001", data["order_number"])
	})

	// تست ۲: سفارش یافت نشد
	t.Run("should return 404 for non-existent order", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/orders/999", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// تست ۳: ID نامعتبر
	t.Run("should return 400 for invalid order ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/orders/invalid", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestOrderHistoryAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uint(1)
	mockOrders := []domain.Order{
		{
			ID:        1,
			UserID:    userID,
			Amount:    100000,
			Status:    domain.OrderCompleted,
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:        2,
			UserID:    userID,
			Amount:    50000,
			Status:    domain.OrderPending,
			CreatedAt: time.Now(),
		},
	}

	mockOrderRepo := new(mocks.MockOrderRepository)
	mockOrderRepo.On("GetHistoryByUserID", mock.Anything, userID).Return(mockOrders, nil)
	mockOrderRepo.On("GetAllOrders", mock.Anything).Return(mockOrders, nil)

	orderService := service.NewOrderService(mockOrderRepo, nil, nil, nil, nil, nil)

	router := gin.New()
	router.GET("/api/v1/users/:id/orders", func(c *gin.Context) {
		orders, _ := orderService.GetAllOrders(c)
		c.JSON(200, orders)
	})

	t.Run("should get user order history", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/users/1/orders", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
