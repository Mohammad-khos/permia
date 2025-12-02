package domain

// AdminInventoryRequest درخواست افزودن موجودی
type AdminInventoryRequest struct {
	ProductSKU string `json:"product_sku" binding:"required"`
	Email      string `json:"email" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Additional string `json:"additional"` // JSON data for tokens, etc.
	MaxUsers   int    `json:"max_users" binding:"min=1"`
	Count      int    `json:"count" binding:"required,min=1"` // تعداد اکانت‌های افزودن شده
}

// AdminInventoryStats آمار موجودی
type AdminInventoryStats struct {
	TotalAccounts     int     `json:"total_accounts"`
	AvailableAccounts int     `json:"available_accounts"`
	SoldAccounts      int     `json:"sold_accounts"`
	ExpiredAccounts   int     `json:"expired_accounts"`
	AvailableRevenue  float64 `json:"available_revenue"`
	ProductSKU        string  `json:"product_sku"`
	Category          string  `json:"category"`
}

// AdminOrderStats آمار سفارشات
type AdminOrderStats struct {
	TotalOrders      int     `json:"total_orders"`
	PendingOrders    int     `json:"pending_orders"`
	PaidOrders       int     `json:"paid_orders"`
	CompletedOrders  int     `json:"completed_orders"`
	FailedOrders     int     `json:"failed_orders"`
	TotalRevenue     float64 `json:"total_revenue"`
	PendingRevenue   float64 `json:"pending_revenue"`
	CompletedRevenue float64 `json:"completed_revenue"`
}
