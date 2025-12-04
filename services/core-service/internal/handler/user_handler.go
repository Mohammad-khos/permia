package handler

import (
	"strconv"

	"Permia/pkg/response"
	"Permia/core-service/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// AuthUser لاگین یا ثبت نام کاربر با تلگرام آیدی
func (h *UserHandler) AuthUser(c *gin.Context) {
	var req struct {
		TelegramID int64  `json:"telegram_id" binding:"required"`
		Username   string `json:"username"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		ReferralCode string `json:"referral_code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid Request Body")
		return
	}

	user, err := h.userSvc.GetOrCreateUser(c, req.TelegramID, req.Username, req.FirstName, req.LastName , req.ReferralCode)
	if err != nil {
		response.ServerError(c, err)
		return
	}

	response.Success(c, user, "User authenticated successfully")
}

// GetBalance دریافت موجودی کیف پول
func (h *UserHandler) GetBalance(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, _ := strconv.Atoi(userIDStr) // در واقعیت باید از Middleware احراز هویت بیاید

	balance, err := h.userSvc.GetBalance(c, uint(userID))
	if err != nil {
		response.Error(c, 404, "User not found")
		return
	}

	response.Success(c, map[string]float64{"balance": balance}, "Balance retrieved")
}

// GetUserByTelegramID دریافت اطلاعات کاربر با تلگرام آیدی
func (h *UserHandler) GetUserByTelegramID(c *gin.Context) {
	telegramIDStr := c.Param("telegram_id")
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		response.Error(c, 400, "Invalid Telegram ID")
		return
	}

	user, err := h.userSvc.GetByTelegramID(c, telegramID)
	if err != nil {
		response.Error(c, 404, "User not found")
		return
	}

	response.Success(c, user, "User profile retrieved")
}