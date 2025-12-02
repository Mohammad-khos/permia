package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	requestURL           = "https://api.zarinpal.com/pg/v4/payment/request.json"
	verifyURL            = "https://api.zarinpal.com/pg/v4/payment/verify.json"
	paymentGatewayURLFmt = "https://www.zarinpal.com/pg/StartPay/%s"
)

// ZarinpalClient is a client for interacting with the Zarinpal API.
type ZarinpalClient struct {
	MerchantID string
	Client     *http.Client
}

// NewZarinpalClient creates a new Zarinpal client.
func NewZarinpalClient() *ZarinpalClient {
	return &ZarinpalClient{
		MerchantID: os.Getenv("ZARINPAL_MERCHANT_ID"),
		Client:     &http.Client{Timeout: 20 * time.Second},
	}
}

// --- Request Structs ---

type paymentRequestBody struct {
	MerchantID  string `json:"merchant_id"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
	CallbackURL string `json:"callback_url"`
}

type verifyRequestBody struct {
	MerchantID string `json:"merchant_id"`
	Amount     int64  `json:"amount"`
	Authority  string `json:"authority"`
}

// --- Response Structs ---

// PaymentResponse is the response from Zarinpal's request endpoint.
type PaymentResponse struct {
	Data struct {
		Authority string `json:"authority"`
		FeeType   string `json:"fee_type"`
		Fee       int    `json:"fee"`
	} `json:"data"`
	Errors []interface{} `json:"errors"`
	// Zarinpal returns code 100 on success
	Code int `json:"code"`
}

// VerificationResponse is the response from Zarinpal's verify endpoint.
type VerificationResponse struct {
	Data struct {
		Code     int    `json:"code"`
		Message  string `json:"message"`
		CardHash string `json:"card_hash"`
		CardPan  string `json:"card_pan"`
		RefID    int64  `json:"ref_id"`
		FeeType  string `json:"fee_type"`
		Fee      int    `json:"fee"`
	} `json:"data"`
	Errors []interface{} `json:"errors"`
	// Zarinpal returns code 100 on success, 101 if already verified
	Code int `json:"code"`
}

// RequestPayment sends a payment request to Zarinpal.
func (c *ZarinpalClient) RequestPayment(amount float64, description, callbackURL string) (*PaymentResponse, error) {
	reqBody := &paymentRequestBody{
		MerchantID:  c.MerchantID,
		Amount:      int64(amount), // Zarinpal uses Toman
		Description: description,
		CallbackURL: callbackURL,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal zarinpal request: %w", err)
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create zarinpal request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to zarinpal: %w", err)
	}
	defer resp.Body.Close()

	var zarinpalResp PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&zarinpalResp); err != nil {
		return nil, fmt.Errorf("failed to decode zarinpal response: %w", err)
	}

	if zarinpalResp.Code != 100 {
		return nil, fmt.Errorf("zarinpal request failed with code %d", zarinpalResp.Code)
	}

	return &zarinpalResp, nil
}

// VerifyPayment verifies a payment with Zarinpal.
func (c *ZarinpalClient) VerifyPayment(amount float64, authority string) (*VerificationResponse, error) {
	reqBody := &verifyRequestBody{
		MerchantID: c.MerchantID,
		Amount:     int64(amount),
		Authority:  authority,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal zarinpal verify request: %w", err)
	}

	req, err := http.NewRequest("POST", verifyURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create zarinpal verify request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send verify request to zarinpal: %w", err)
	}
	defer resp.Body.Close()

	var zarinpalResp VerificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&zarinpalResp); err != nil {
		return nil, fmt.Errorf("failed to decode zarinpal verify response: %w", err)
	}

	return &zarinpalResp, nil
}

// GetPaymentURL returns the URL for the user to proceed with payment.
func GetPaymentURL(authority string) string {
	return fmt.Sprintf(paymentGatewayURLFmt, authority)
}
