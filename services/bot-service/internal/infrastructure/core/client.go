package core

import (
	"Permia/bot-service/internal/domain"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type Client struct {
	resty  *resty.Client
	logger *zap.SugaredLogger
}

func NewClient(baseURL string, logger *zap.SugaredLogger) *Client {
	r := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json")

	return &Client{
		resty:  r,
		logger: logger,
	}
}

// LoginUser calls the core service to log in or register a user.
func (c *Client) LoginUser(telegramID int64, username, firstName, lastName string) (*domain.User, error) {
	var user domain.User
	payload := map[string]interface{}{
		"telegram_id": telegramID,
		"username":    username,
		"first_name":  firstName,
		"last_name":   lastName,
	}

	resp, err := c.resty.R().
		SetBody(payload).
		SetResult(&user).
		Post("/users/auth")

	if err != nil {
		c.logger.Errorf("Core service LoginUser request failed: %v", err)
		return nil, fmt.Errorf("core service is unavailable")
	}
	if resp.IsError() {
		c.logger.Errorf("Core service LoginUser returned error: %s", resp.String())
		return nil, fmt.Errorf("failed to login user, status: %s", resp.Status())
	}

	return &user, nil
}

// GetProfile calls the core service to get a user's profile.
func (c *Client) GetProfile(telegramID int64) (*domain.User, error) {
	var result struct {
		Data domain.User `json:"data"`
	}

	// استفاده از اندپوینت جدید برای دریافت دیتای کامل
	resp, err := c.resty.R().
		SetResult(&result).
		Get(fmt.Sprintf("/users/by-telegram/%d", telegramID))

	if err != nil {
		c.logger.Errorf("GetProfile request failed: %v", err)
		return nil, fmt.Errorf("core service is unavailable")
	}
	if resp.StatusCode() == 404 {
		return nil, nil
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to get profile: %s", resp.Status())
	}

	return &result.Data, nil
}

// GetProducts calls the core service to get the list of available products.
func (c *Client) GetProducts() ([]domain.Product, error) {
	var res struct {
		Success bool                        `json:"success"`
		Message string                      `json:"message"`
		Data    map[string][]domain.Product `json:"data"`
	}

	resp, err := c.resty.R().
		SetResult(&res).
		Get("/products")

	if err != nil {
		c.logger.Errorf("Core service GetProducts request failed: %v", err)
		return nil, fmt.Errorf("core service is unavailable")
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to get products, status: %s", resp.Status())
	}

	var products []domain.Product
	for _, list := range res.Data {
		products = append(products, list...)
	}

	return products, nil
}

// CreateOrderRequest ساختار آپدیت شده (شامل TelegramID)
type CreateOrderRequest struct {
	UserID     uint   `json:"user_id"`
	TelegramID int64  `json:"telegram_id"` // فیلد ضروری جدید
	SKU        string `json:"sku"`
}

type CreateOrderResponse struct {
	OrderID       uint    `json:"id"`
	OrderNumber   string  `json:"order_number"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	DeliveredData string  `json:"delivered_data"`
}

// CreateOrder متد آپدیت شده: دریافت telegramID به عنوان پارامتر دوم
func (c *Client) CreateOrder(userID uint, telegramID int64, sku string) (*CreateOrderResponse, error) {
	payload := CreateOrderRequest{
		UserID:     userID,
		TelegramID: telegramID,
		SKU:        sku,
	}

	var result CreateOrderResponse
	resp, err := c.resty.R().
		SetBody(payload).
		SetResult(&result).
		Post("/orders")

	if err != nil {
		c.logger.Errorf("CreateOrder request failed: %v", err)
		return nil, fmt.Errorf("core service is unavailable")
	}
	
	// بررسی دقیق‌تر خطا
	if resp.IsError() {
		c.logger.Errorf("Core CreateOrder error body: %s", resp.String())
		if resp.StatusCode() == 400 && contains(resp.String(), "insufficient") {
			return nil, fmt.Errorf("insufficient funds")
		}
		// برگرداندن متن خطا برای دیباگ بهتر
		return nil, fmt.Errorf("failed to create order: %s", resp.String())
	}

	return &result, nil
}

// GetUserSubscriptions دریافت لیست اشتراک‌ها
func (c *Client) GetUserSubscriptions(telegramID int64) ([]domain.Subscription, error) {
	var response struct {
		Data []domain.Subscription `json:"data"`
	}

	resp, err := c.resty.R().
		SetHeader("X-Telegram-ID", fmt.Sprintf("%d", telegramID)).
		SetResult(&response).
		Get("/users/subscriptions")

	if err != nil || resp.IsError() {
		return nil, err
	}

	return response.Data, nil
}

// GetPaymentLink دریافت لینک پرداخت
func (c *Client) GetPaymentLink(userID uint, amount float64) (string, error) {
	var result struct {
		VerificationURL string `json:"verification_url"`
	}
	payload := map[string]interface{}{
		"user_id":        userID,
		"amount":         amount,
		"payment_method": "card",
	}

	resp, err := c.resty.R().
		SetBody(payload).
		SetResult(&result).
		Post("/payment/charge")

	if err != nil || resp.IsError() {
		return "", fmt.Errorf("payment service unavailable")
	}

	return result.VerificationURL, nil
}

func contains(str, substr string) bool {
	return strings.Contains(str, substr)
}