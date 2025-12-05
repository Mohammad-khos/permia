package handler

import (
	"Permia/core-service/internal/service"
	"Permia/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CouponHandler struct {
	couponSvc *service.CouponService
	userSvc   *service.UserService // برای تبدیل telegram_id به user_id
}

func NewCouponHandler(couponSvc *service.CouponService, userSvc *service.UserService) *CouponHandler {
	return &CouponHandler{couponSvc: couponSvc, userSvc: userSvc}
}

type ValidateRequest struct {
	TelegramID int64   `json:"telegram_id" binding:"required"`
	Code       string  `json:"code" binding:"required"`
	Amount     float64 `json:"amount" binding:"required"`
}

func (h *CouponHandler) Validate(c *gin.Context) {
	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request")
		return
	}

	user, err := h.userSvc.GetByTelegramID(c, req.TelegramID)
	if err != nil || user == nil {
		response.Error(c, 404, "user not found")
		return
	}

	finalPrice, discount, err := h.couponSvc.ValidateCoupon(c, req.Code, user.ID, req.Amount)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, gin.H{
		"final_price":     finalPrice,
		"discount_amount": discount,
		"valid":           true,
	}, "coupon valid")
}

func (h *CouponHandler) GetMyCoupons(c *gin.Context) {
	telegramID, _ := strconv.ParseInt(c.Param("telegram_id"), 10, 64)
	user, err := h.userSvc.GetByTelegramID(c, telegramID)
	if err != nil || user == nil {
		response.Error(c, 404, "user not found")
		return
	}

	coupons, err := h.couponSvc.GetUserCoupons(c, user.ID)
	if err != nil {
		response.ServerError(c, err)
		return
	}

	response.Success(c, coupons, "coupons retrieved")
}