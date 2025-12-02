package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Permia/core-service/internal/domain"
	"Permia/core-service/internal/handler"
	"Permia/core-service/internal/service"
	"Permia/core-service/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListProductsAPI(t *testing.T) {
	// 1. تنظیم جین روی حالت تست (برای خاموش کردن لاگ‌های اضافه)
	gin.SetMode(gin.TestMode)

	// 2. آماده‌سازی داده‌های فیک (Fake Data)
	mockProducts := []domain.Product{
		{
			SKU:      "gpt_test",
			Category: "chatgpt",
			Title:    "Test GPT Product",
			Price:    1000,
			IsActive: true,
		},
		{
			SKU:      "gemini_test",
			Category: "gemini",
			Title:    "Test Gemini Product",
			Price:    2000,
			IsActive: true,
		},
	}

	// 3. ساخت Mock و تزریق وابستگی‌ها
	mockRepo := new(mocks.MockProductRepository)
	
	// سناریو: وقتی متد GetActiveProducts صدا زده شد، لیست بالا را برگردان
	mockRepo.On("GetActiveProducts", mock.Anything).Return(mockProducts, nil)

	// ساخت سرویس واقعی با ریپازیتوری فیک
	productService := service.NewProductService(mockRepo)
	
	// ساخت هندلر واقعی با سرویس
	productHandler := handler.NewProductHandler(productService)

	// 4. راه اندازی روتر Gin
	r := gin.New()
	r.GET("/api/v1/products", productHandler.ListProducts)

	// 5. شبیه‌سازی ریکوئست HTTP
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/products", nil)
	w := httptest.NewRecorder() // ضبط کننده پاسخ

	// 6. اجرا
	r.ServeHTTP(w, req)

	// 7. بررسی نتایج (Assertions)
	assert.Equal(t, http.StatusOK, w.Code) // باید 200 باشد

	// پارس کردن بادی ریسپانس
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"]) // فیلد success باید true باشد
	
	// بررسی اینکه دیتا درست برگشته (باید شامل دسته‌بندی chatgpt و gemini باشد)
	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["chatgpt"])
	assert.NotNil(t, data["gemini"])

	// اطمینان از اینکه متد ماک صدا زده شده
	mockRepo.AssertExpectations(t)
}