package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response ساختار استاندارد خروجی جیسون
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success پاسخ موفقیت‌آمیز 200
func Success(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created پاسخ ساخت موفق 201
func Created(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error پاسخ خطا (کد وضعیت قابل تنظیم)
func Error(c *gin.Context, statusCode int, errMessage string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   errMessage,
	})
}

// ServerError پاسخ خطای سرور 500
func ServerError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error:   "Internal Server Error",
		Message: err.Error(), // در پروداکشن واقعی شاید نخواهیم جزئیات خطا را بفرستیم
	})
}