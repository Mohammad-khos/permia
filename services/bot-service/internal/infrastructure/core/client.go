package core

import (
	"Permia/bot-service/internal/domain"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// Client is a client for the core service API.
type Client struct {
	resty  *resty.Client
	logger *zap.SugaredLogger
}

// NewClient creates a new core service client.
func NewClient(baseURL string, logger *zap.SugaredLogger) *Client {
	r := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json")

	return &Client{
		resty:  r,
		logger: logger,
	}
}

// LoginUserRequest is the payload for the login request.
type LoginUserRequest struct {
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
}

// LoginUser calls the core service to log in or register a user.
func (c *Client) LoginUser(telegramID int64, username, firstName, lastName string) (*domain.User, error) {
	var user domain.User
	payload := LoginUserRequest{
		TelegramID: telegramID,
		Username:   username,
		FirstName:  firstName,
		LastName:   lastName,
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

// GetProfile calls the core service to get a user's profile and balance.
func (c *Client) GetProfile(telegramID int64) (*domain.User, error) {
	var user domain.User
	resp, err := c.resty.R().
		SetResult(&user).
		Get(fmt.Sprintf("/users/by-telegram/%d/balance", telegramID))

	if err != nil {
		c.logger.Errorf("Core service GetProfile request failed: %v", err)
		return nil, fmt.Errorf("core service is unavailable")
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, nil // User not found is not a system error
	}
	if resp.IsError() {
		c.logger.Errorf("Core service GetProfile returned error: %s", resp.String())
		return nil, fmt.Errorf("failed to get profile, status: %s", resp.Status())
	}

	return &user, nil
}

// GetProducts calls the core service to get the list of available products.
func (c *Client) GetProducts() ([]domain.Product, error) {
	// Core returns a standard response wrapping data under `data`.
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
		c.logger.Errorf("Core service GetProducts returned error: %s", resp.String())
		return nil, fmt.Errorf("failed to get products, status: %s", resp.Status())
	}

	// Flatten catalog into slice
	var products []domain.Product
	for _, list := range res.Data {
		products = append(products, list...)
	}

	return products, nil
}

// CreateOrderRequest is the payload for creating an order
type CreateOrderRequest struct {
	UserID    uint   `json:"user_id"`
	ProductID uint   `json:"product_id"`
	SKU       string `json:"sku"`
}

// CreateOrderResponse is the response from creating an order
type CreateOrderResponse struct {
	OrderID       uint    `json:"id"`
	OrderNumber   string  `json:"order_number"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	DeliveredData string  `json:"delivered_data"`
}

// CreateOrder creates a new order
func (c *Client) CreateOrder(userID uint, sku string) (*CreateOrderResponse, error) {
	payload := CreateOrderRequest{
		UserID: userID,
		SKU:    sku,
	}

	var result CreateOrderResponse
	resp, err := c.resty.R().
		SetBody(payload).
		SetResult(&result).
		Post("/orders")

	if err != nil {
		c.logger.Errorf("Core service CreateOrder request failed: %v", err)
		return nil, fmt.Errorf("core service is unavailable")
	}
	if resp.IsError() {
		c.logger.Errorf("Core service CreateOrder returned error: %s", resp.String())
		// Check for insufficient funds error
		if resp.StatusCode() == 400 && contains(resp.String(), "insufficient") {
			return nil, fmt.Errorf("insufficient funds")
		}
		return nil, fmt.Errorf("failed to create order: %s", resp.String())
	}

	return &result, nil
}

// ChargeRequest is the payload for initiating a charge
type ChargeRequest struct {
	OrderID       uint   `json:"order_id"`
	UserID        uint   `json:"user_id"`
	PaymentMethod string `json:"payment_method"`
}

// ChargeResponse is the response from charging payment
type ChargeResponse struct {
	PaymentID       uint    `json:"payment_id"`
	OrderID         uint    `json:"order_id"`
	Amount          float64 `json:"amount"`
	Status          string  `json:"status"`
	VerificationURL string  `json:"verification_url"`
	TransactionID   string  `json:"transaction_id"`
}

// GetPaymentLink initiates a payment charge and returns the payment link
func (c *Client) GetPaymentLink(userID uint, amount float64) (string, error) {
	// First create a charge request
	// Note: In production, you'd typically create an order first
	payload := map[string]interface{}{
		"user_id":        userID,
		"amount":         amount,
		"payment_method": "card",
	}

	var result ChargeResponse
	resp, err := c.resty.R().
		SetBody(payload).
		SetResult(&result).
		Post("/payment/charge")

	if err != nil {
		c.logger.Errorf("Core service GetPaymentLink request failed: %v", err)
		return "", fmt.Errorf("payment service unavailable")
	}
	if resp.IsError() {
		c.logger.Errorf("Core service GetPaymentLink returned error: %s", resp.String())
		return "", fmt.Errorf("failed to get payment link: %s", resp.String())
	}

	return result.VerificationURL, nil
}

// helper function to check if string contains substring
func contains(str, substr string) bool {
	return strings.Contains(str, substr)
}
