package handler

import (
	"strconv"
	"strings"

	"Permia/core-service/internal/domain"
	"Permia/core-service/internal/service"
	"Permia/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	adminSvc   *service.AdminService
	productSvc *service.ProductService
}

func NewAdminHandler(adminSvc *service.AdminService, productSvc *service.ProductService) *AdminHandler {
	return &AdminHandler{
		adminSvc:   adminSvc,
		productSvc: productSvc,
	}
}

// CreateInventory افزودن موجودی جدید
// POST /api/v1/admin/inventory
func (h *AdminHandler) CreateInventory(c *gin.Context) {
	var req domain.AdminInventoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid request body")
		return
	}

	result, err := h.adminSvc.AddInventory(c, &req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, result, "Inventory added successfully")
}

// GetInventoryStats دریافت آمار موجودی
// GET /api/v1/admin/inventory/stats?sku=gpt_shared_4
func (h *AdminHandler) GetInventoryStats(c *gin.Context) {
	sku := c.Query("sku")
	if sku == "" {
		response.Error(c, 400, "SKU parameter is required")
		return
	}

	stats, err := h.adminSvc.GetInventoryStats(c, sku)
	if err != nil {
		response.ServerError(c, err)
		return
	}

	response.Success(c, stats, "Inventory stats retrieved successfully")
}

// GetOrderStats دریافت آمار سفارشات
// GET /api/v1/admin/orders
func (h *AdminHandler) GetOrderStats(c *gin.Context) {
	stats, err := h.adminSvc.GetOrderStats(c)
	if err != nil {
		response.ServerError(c, err)
		return
	}

	response.Success(c, stats, "Order stats retrieved successfully")
}

// CompleteOrder تکمیل سفارش توسط ادمین
// POST /api/v1/admin/orders/:id/complete
func (h *AdminHandler) CompleteOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, 400, "Invalid order ID")
		return
	}

	result, err := h.adminSvc.CompleteOrder(c, uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "سفارش پیدا نشد") {
			response.Error(c, 404, err.Error())
			return
		}
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, result, "Order completed successfully")
}

// CreateProduct ایجاد محصول جدید
// POST /api/v1/admin/products
func (h *AdminHandler) CreateProduct(c *gin.Context) {
	var req struct {
		SKU          string  `json:"sku" binding:"required"`
		Category     string  `json:"category" binding:"required"`
		Title        string  `json:"title" binding:"required"`
		Description  string  `json:"description"`
		Price        float64 `json:"price" binding:"required,gt=0"`
		Type         string  `json:"type" binding:"required"`
		Capacity     int     `json:"capacity" binding:"min=1"`
		DisplayOrder int     `json:"display_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid request body")
		return
	}

	product := &domain.Product{
		SKU:          req.SKU,
		Category:     req.Category,
		Title:        req.Title,
		Description:  req.Description,
		Price:        req.Price,
		Type:         req.Type,
		Capacity:     req.Capacity,
		IsActive:     true,
		DisplayOrder: req.DisplayOrder,
	}

	// ذخیره‌سازی محصول
	// در یک سرویس واقعی، این کار را یک متد سرویس انجام می‌دهد
	dbVal := c.Request.Context().Value("db")
	if dbVal == nil {
		response.Error(c, 500, "Database connection not available")
		return
	}

	db := dbVal.(*gorm.DB)
	if err := db.Create(product).Error; err != nil {
		response.Error(c, 400, "Failed to create product")
		return
	}

	response.Success(c, product, "Product created successfully")
}

// UpdateProduct به‌روز‌رسانی محصول
// PUT /api/v1/admin/products/:sku
func (h *AdminHandler) UpdateProduct(c *gin.Context) {
	sku := c.Param("sku")
	if sku == "" {
		response.Error(c, 400, "SKU parameter is required")
		return
	}

	var req struct {
		Title        string  `json:"title"`
		Description  string  `json:"description"`
		Price        float64 `json:"price" binding:"omitempty,gt=0"`
		IsActive     *bool   `json:"is_active"`
		DisplayOrder *int    `json:"display_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid request body")
		return
	}

	// بروزرسانی محصول
	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Price > 0 {
		updates["price"] = req.Price
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.DisplayOrder != nil {
		updates["display_order"] = *req.DisplayOrder
	}

	dbVal := c.Request.Context().Value("db")
	if dbVal == nil {
		response.Error(c, 500, "Database connection not available")
		return
	}

	db := dbVal.(*gorm.DB)
	if err := db.Model(&domain.Product{}).Where("sku = ?", sku).Updates(updates).Error; err != nil {
		response.Error(c, 400, "Failed to update product")
		return
	}

	product, _ := h.productSvc.GetProductBySKU(c, sku)
	response.Success(c, product, "Product updated successfully")
}
