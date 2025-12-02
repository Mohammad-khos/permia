package brocard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"Permia/core-service/internal/domain"
)

type BrocardClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

func NewBrocardClient() *BrocardClient {
	return &BrocardClient{
		// آدرس استاندارد API بروکارد
		BaseURL: "https://private.mybrocard.com/api/v1",
		Token:   os.Getenv("BROCARD_API_TOKEN"),
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// IssueCard درخواست صدور کارت
func (c *BrocardClient) IssueCard(amount float64, cardTypeID string) (*domain.VirtualCard, error) {
	// بدنه درخواست طبق داکیومنت بروکارد
	reqBody := map[string]interface{}{
		"type":    cardTypeID, // مثلا "visa_universal" یا آیدی عددی
		"balance": amount,
		"comment": "Permia Order",
	}

	respBytes, err := c.postRequest("/cards/issue", reqBody)
	if err != nil {
		return nil, err
	}

	// پارس کردن پاسخ (ساختار احتمالی بروکارد)
	var apiRes struct {
		Success bool `json:"success"`
		Data    struct {
			ID     int    `json:"id"`
			Pan    string `json:"card_number"`
			Cvv    string `json:"cvv"`
			Expiry string `json:"expiry"`
		} `json:"data"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(respBytes, &apiRes); err != nil {
		return nil, fmt.Errorf("parse error: %v", err)
	}

	if !apiRes.Success {
		return nil, fmt.Errorf("brocard error: %s", apiRes.Message)
	}

	return &domain.VirtualCard{
		ID:     fmt.Sprintf("%d", apiRes.Data.ID),
		PAN:    apiRes.Data.Pan,
		CVV:    apiRes.Data.Cvv,
		Expiry: apiRes.Data.Expiry,
	}, nil
}

func (c *BrocardClient) GetBalance() (float64, error) {
	// فعلاً پیاده‌سازی نشده (مهم نیست برای شروع)
	return 0, nil
}

// postRequest تابع کمکی ارسال درخواست
func (c *BrocardClient) postRequest(endpoint string, body interface{}) ([]byte, error) {
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", c.BaseURL+endpoint, bytes.NewBuffer(jsonBody))
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// خواندن بافر پاسخ
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	return buf.Bytes(), nil
}