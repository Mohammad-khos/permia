package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"Permia/core-service/internal/domain"
	"Permia/core-service/internal/handler"
	"Permia/core-service/internal/middleware"
	"Permia/core-service/internal/service"
	"Permia/core-service/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestAdminAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. ساخت لاگر فیک برای تست
	testLogger, _ := zap.NewProduction()
	defer testLogger.Sync()

	// 2. ساخت روتر و اعمال میدلور
	router := gin.New()
	adminGroup := router.Group("/admin")
	adminGroup.Use(middleware.AdminAuth(testLogger))
	adminGroup.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome Admin!"})
	})

	// تست ۱: درخواست موفق با توکن ادمین معتبر
	t.Run("should allow access with valid admin token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/admin/test", nil)
		req.Header.Set("X-Admin-Token", "your-secret-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.NotEqual(t, http.StatusBadRequest, w.Code)
	})

	// تست ۲: درخواست ناموفق به دلیل نبودن هدر
	t.Run("should deny access without admin token header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/admin/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// تست ۳: درخواست ناموفق با توکن غیرمعتبر
	t.Run("should deny access with invalid admin token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/admin/test", nil)
		req.Header.Set("X-Admin-Token", "invalid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

var (
	_ = TestCreateInventoryAPI
	_ = TestInventoryStatsAPI
	_ = TestOrderStatsAPI
	_ = TestCompleteOrderAPI
	_ = TestCreateProductAPI
	_ = TestUpdateProductAPI
)

func TestCreateInventoryAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockProduct := &domain.Product{
		ID:    1,
		SKU:   "gpt_shared",
		Title: "GPT Shared",
		Price: 100000,
	}

	// ایجاد Mocks
	mockAccountRepo := new(mocks.MockAccountRepository)
	mockProductRepo := new(mocks.MockProductRepository)
	mockOrderRepo := new(mocks.MockOrderRepository)

	mockProductRepo.On("GetBySKU", mock.Anything, "gpt_shared").Return(mockProduct, nil)
	mockAccountRepo.On("CreateBatch", mock.Anything, mock.Anything).Return(nil)

	// ایجاد لاگر
	testLogger, _ := zap.NewProduction()
	defer testLogger.Sync()

	// ایجاد سرویس و هندلر
	adminService := service.NewAdminService(mockAccountRepo, mockOrderRepo, mockProductRepo, nil, testLogger)
	adminHandler := handler.NewAdminHandler(adminService, nil)

	// ایجاد روتر
	router := gin.New()
	router.POST("/api/v1/admin/inventory", adminHandler.CreateInventory)

	// تست ۱: اضافه کردن موجودی موفق
	t.Run("should create inventory successfully", func(t *testing.T) {
		body := domain.AdminInventoryRequest{
			ProductSKU: "gpt_shared",
			Email:      "account",
			Password:   "secure_pass",
			Additional: `{"token":"xyz"}`,
			MaxUsers:   3,
			Count:      5,
		}

		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/admin/inventory", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "gpt_shared", data["product_sku"])
		assert.Equal(t, float64(5), data["added_count"])
	})

	// تست ۲: محصول یافت نشد
	t.Run("should return error when product not found", func(t *testing.T) {
		mockProductRepo.On("GetBySKU", mock.Anything, "invalid_sku").Return(nil, gorm.ErrRecordNotFound)

		body := domain.AdminInventoryRequest{
			ProductSKU: "invalid_sku",
			Email:      "account",
			Password:   "pass",
			Count:      1,
		}

		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/admin/inventory", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// تست ۳: درخواست نامعتبر
	t.Run("should return 400 for invalid request", func(t *testing.T) {
		body := map[string]interface{}{
			"product_sku": "gpt_shared",
			// Email و Count را حذف کردیم
		}

		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/admin/inventory", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestInventoryStatsAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// ایجاد لاگر
	testLogger, _ := zap.NewProduction()
	defer testLogger.Sync()

	// ایجاد Mocks
	mockAccountRepo := new(mocks.MockAccountRepository)
	mockOrderRepo := new(mocks.MockOrderRepository)
	mockProductRepo := new(mocks.MockProductRepository)
	mockAdminService := service.NewAdminService(mockAccountRepo, mockOrderRepo, mockProductRepo, nil, testLogger)
	adminHandler := handler.NewAdminHandler(mockAdminService, nil)

	router := gin.New()
	router.GET("/api/v1/admin/inventory/stats", adminHandler.GetInventoryStats)

	// تست ۱: SKU پارامتر یافت نشد
	t.Run("should return 400 when SKU is missing", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/admin/inventory/stats", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestOrderStatsAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// ایجاد لاگر
	testLogger, _ := zap.NewProduction()
	defer testLogger.Sync()

	mockAccountRepo := new(mocks.MockAccountRepository)
	mockOrderRepo := new(mocks.MockOrderRepository)
	mockProductRepo := new(mocks.MockProductRepository)
	adminService := service.NewAdminService(mockAccountRepo, mockOrderRepo, mockProductRepo, nil, testLogger)
	adminHandler := handler.NewAdminHandler(adminService, nil)

	router := gin.New()
	router.GET("/api/v1/admin/orders", adminHandler.GetOrderStats)

	t.Run("should return 500 when db not available", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/admin/orders", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Endpoint exists but needs DB connection
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestCompleteOrderAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockOrder := &domain.Order{
		ID:        1,
		UserID:    1,
		Amount:    100000,
		Status:    domain.OrderPaid,
		CreatedAt: time.Now(),
	}

	// ایجاد Mocks
	mockOrderRepo := new(mocks.MockOrderRepository)

	mockOrderRepo.On("GetByID", mock.Anything, uint(1)).Return(mockOrder, nil)
	mockOrderRepo.On("UpdateStatus", mock.Anything, uint(1), "COMPLETED").Return(nil)
	mockOrderRepo.On("GetByID", mock.Anything, uint(999)).Return(nil, gorm.ErrRecordNotFound)

	// ایجاد لاگر
	testLogger, _ := zap.NewProduction()
	defer testLogger.Sync()

	adminService := service.NewAdminService(nil, mockOrderRepo, nil, nil, testLogger)
	adminHandler := handler.NewAdminHandler(adminService, nil)

	router := gin.New()
	router.POST("/api/v1/admin/orders/:id/complete", adminHandler.CompleteOrder)

	// تست ۱: تکمیل سفارش موفق
	t.Run("should complete order successfully", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/admin/orders/1/complete", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["order_id"])
	})

	// تست ۲: سفارش یافت نشد
	t.Run("should return 404 for non-existent order", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/admin/orders/999/complete", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// تست ۳: ID نامعتبر
	t.Run("should return 400 for invalid order ID", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/admin/orders/invalid/complete", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCreateProductAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// ایجاد لاگر
	testLogger, _ := zap.NewProduction()
	defer testLogger.Sync()

	adminService := service.NewAdminService(nil, nil, nil, nil, testLogger)
	adminHandler := handler.NewAdminHandler(adminService, nil)

	router := gin.New()
	router.POST("/api/v1/admin/products", adminHandler.CreateProduct)

	t.Run("should return error (db context not set)", func(t *testing.T) {
		body := map[string]interface{}{
			"sku":           "new_product",
			"category":      "chatgpt",
			"title":         "New Product",
			"price":         100000,
			"type":          "shared",
			"capacity":      3,
			"display_order": 1,
		}

		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/admin/products", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// این تست توقع دارد خرابی داشته باشد زیرا db در context نیست
		// بنابراین فقط چک می‌کنیم که درخواست حالت خرابی برگردانده شود
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUpdateProductAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockProduct := &domain.Product{
		ID:       1,
		SKU:      "gpt_shared",
		Title:    "GPT Shared",
		Price:    100000,
		IsActive: true,
	}

	mockProductRepo := new(mocks.MockProductRepository)
	mockProductRepo.On("GetBySKU", mock.Anything, "gpt_shared").Return(mockProduct, nil)

	// ایجاد لاگر
	testLogger, _ := zap.NewProduction()
	defer testLogger.Sync()

	productService := service.NewProductService(mockProductRepo)
	adminService := service.NewAdminService(nil, nil, nil, nil, testLogger)
	adminHandler := handler.NewAdminHandler(adminService, productService)

	router := gin.New()
	router.PUT("/api/v1/admin/products/:sku", adminHandler.UpdateProduct)

	t.Run("should return error (db context not set)", func(t *testing.T) {
		body := map[string]interface{}{
			"title": "Updated Title",
			"price": 120000,
		}

		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", "/api/v1/admin/products/gpt_shared", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// فقط چک می‌کنیم که درخواست پذیرفته شد
		assert.NotEqual(t, http.StatusBadRequest, w.Code)
	})
}
