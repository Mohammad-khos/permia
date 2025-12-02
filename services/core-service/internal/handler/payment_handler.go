package handler

import (
	"Permia/core-service/internal/service"
	"Permia/pkg/response"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentSvc *service.PaymentService
}

func NewPaymentHandler(paymentSvc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentSvc: paymentSvc,
	}
}

// ChargePayment شروع فرآیند پرداخت
// POST /api/v1/payment/charge
func (h *PaymentHandler) ChargePayment(c *gin.Context) {
	var req struct {
		OrderID       uint   `json:"order_id" binding:"required"`
		UserID        uint   `json:"user_id" binding:"required"`
		PaymentMethod string `json:"payment_method" binding:"required,oneof=card wallet"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid request body")
		return
	}

	chargeReq := &service.ChargeRequest{
		OrderID:       req.OrderID,
		UserID:        req.UserID,
		PaymentMethod: req.PaymentMethod,
	}

	result, err := h.paymentSvc.Charge(c, chargeReq)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, result, "Payment initiated successfully")
}

// VerifyPayment تأیید پرداخت
// GET /api/v1/payment/verify?payment_id=1&authority=ABC123
func (h *PaymentHandler) VerifyPayment(c *gin.Context) {
	var req struct {
		PaymentID uint   `form:"payment_id" binding:"required"`
		Authority string `form:"authority" binding:"required"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, 400, "Invalid query parameters")
		return
	}

	verifyReq := &service.VerifyRequest{
		PaymentID: req.PaymentID,
		Authority: req.Authority,
	}

	result, err := h.paymentSvc.Verify(c, verifyReq)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, result, "Payment verification completed")
}
