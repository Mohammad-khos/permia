package handler

import (
	"Permia/core-service/internal/service"
	"Permia/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderSvc *service.OrderService
	userSvc  *service.UserService // برای چک کردن یوزر قبل از خرید
}

func NewOrderHandler(orderSvc *service.OrderService, userSvc *service.UserService) *OrderHandler {
	return &OrderHandler{
		orderSvc: orderSvc,
		userSvc:  userSvc,
	}
}

// CreateOrder ثبت سفارش جدید
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req struct {
		TelegramID int64  `json:"telegram_id" binding:"required"`
		ProductSKU string `json:"sku" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid Request Body")
		return
	}

	// 1. پیدا کردن کاربر (چون شاید هنوز تو دیتابیس کش نشده باشه یا درخواست مستقیم باشه)
	// نکته: در معماری میکروسرویس بهتره UserID رو از توکن بگیریم، ولی اینجا با TelegramID کار میکنیم
	// فرض میکنیم کاربر قبلا Auth شده و UserID رو داریم، یا همینجا سریع فچ میکنیم
	user, err := h.userSvc.GetOrCreateUser(c, req.TelegramID, "", "", "")
	if err != nil {
		response.ServerError(c, err)
		return
	}

	// 2. انجام خرید
	result, err := h.orderSvc.PurchaseFlow(c, user.ID, req.ProductSKU)
	if err != nil {
		// ارورهای بیزینسی (موجودی کم و...) رو با کد 400 برمیگردونیم
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, result, "Order created successfully")
}

// GetOrders لیست تمام سفارشات (برای مدیریت)
// GET /api/v1/orders
func (h *OrderHandler) GetOrders(c *gin.Context) {
	orders, err := h.orderSvc.GetAllOrders(c)
	if err != nil {
		response.ServerError(c, err)
		return
	}

	response.Success(c, orders, "Orders retrieved successfully")
}

// GetOrderByID دریافت جزئیات یک سفارش
// GET /api/v1/orders/:id
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, 400, "Invalid order ID")
		return
	}

	order, err := h.orderSvc.GetOrderByID(c, uint(id))
	if err != nil {
		response.Error(c, 404, "Order not found")
		return
	}

	response.Success(c, order, "Order retrieved successfully")
}
