package handler

import (
	"Permia/core-service/internal/service"
	"Permia/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderSvc *service.OrderService
	userSvc  *service.UserService
}

// SubscriptionResponse ساختار پاسخ سفارش اشتراک
type SubscriptionResponse struct {
	ID            uint    `json:"id"`
	ProductName   string  `json:"ProductName"`
	Sku           string  `json:"sku"`
	DeliveredData string  `json:"DeliveredData"`
	Amount        float64 `json:"amount"`
	CreatedAt     string  `json:"CreatedAt"`
	ExpiresAt     string  `json:"ExpiresAt"`
}

// CreateOrderRequest ساختار ورودی منعطف برای ثبت سفارش
type CreateOrderRequest struct {
	UserID     uint   `json:"user_id"`      // ارسالی از سمت بات
	TelegramID int64  `json:"telegram_id"`  // برای پشتیبانی از کدهای قدیمی
	SKU        string `json:"sku" binding:"required"`
	CouponCode string `json:"coupon_code"`
}

func NewOrderHandler(orderSvc *service.OrderService, userSvc *service.UserService) *OrderHandler {
	return &OrderHandler{
		orderSvc: orderSvc,
		userSvc:  userSvc,
	}
}

// CreateOrder ثبت سفارش جدید
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest

	// 1. دریافت و اعتبارسنجی ورودی
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid Request Body: " + err.Error())
		return
	}

	// 2. تعیین شناسه کاربر (UserID)
	var finalUserID uint

	if req.UserID > 0 {
		// اگر بات شناسه داخلی کاربر را فرستاده بود، از همان استفاده می‌کنیم
		finalUserID = req.UserID
	} else if req.TelegramID > 0 {
		// اگر فقط تلگرام آیدی داشتیم، کاربر را پیدا یا ایجاد می‌کنیم
		user, err := h.userSvc.GetOrCreateUser(c, req.TelegramID, "", "", "" , "")
		if err != nil {
			c.Error(err)
			response.ServerError(c, err)
			return
		}
		finalUserID = user.ID
	} else {
		response.Error(c, 400, "Both user_id and telegram_id are missing")
		return
	}

	// 3. انجام خرید با شناسه نهایی کاربر
	result, err := h.orderSvc.PurchaseFlow(c, finalUserID, req.SKU , req.CouponCode)
	if err != nil {
		// خطا را چک می‌کنیم تا اگر مربوط به موجودی بود، کد مناسب برگردانیم
		if err.Error() == "موجودی کافی نیست" || err.Error() == "insufficient funds" {
			response.Error(c, 400, "insufficient funds")
			return
		}
		// سایر ارورهای بیزینسی
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, result, "Order created successfully")
}

// GetOrders لیست تمام سفارشات (برای مدیریت)
func (h *OrderHandler) GetOrders(c *gin.Context) {
	orders, err := h.orderSvc.GetAllOrders(c)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.Success(c, orders, "Orders retrieved successfully")
}

// GetOrderByID دریافت جزئیات یک سفارش
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

// GetUserSubscriptions دریافت سفارشات اشتراک کاربر
func (h *OrderHandler) GetUserSubscriptions(c *gin.Context) {
	telegramIDStr := c.GetHeader("X-Telegram-ID")
	if telegramIDStr == "" {
		response.Error(c, 400, "X-Telegram-ID header is required")
		return
	}

	telegramID, _ := strconv.ParseInt(telegramIDStr, 10, 64)

	orders, err := h.orderSvc.GetUserSubscriptions(c, telegramID)
	if err != nil {
		response.Error(c, 404, "User or orders not found")
		return
	}

	var result []SubscriptionResponse
	for _, o := range orders {
		days := 30
		expiryTime := o.CreatedAt.AddDate(0, 0, days)

		result = append(result, SubscriptionResponse{
			ID:            o.ID,
			ProductName:   o.Product.Title,
			Sku:           o.Product.SKU,
			DeliveredData: o.DeliveredData,
			Amount:        o.Amount,
			CreatedAt:     o.CreatedAt.Format("2006-01-02"),
			ExpiresAt:     expiryTime.Format("2006-01-02"),
		})
	}

	response.Success(c, result, "User subscriptions retrieved")
}