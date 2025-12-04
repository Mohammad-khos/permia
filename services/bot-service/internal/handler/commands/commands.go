package commands

import (
	"Permia/bot-service/internal/handler/menus"
	"Permia/bot-service/internal/service"
	"fmt"
	"strings"

	"gopkg.in/telebot.v3"
)

type Handler struct {
	botService *service.BotService
}

func NewHandler(botService *service.BotService) *Handler {
	return &Handler{botService: botService}
}

// Start handles the /start command.
func (h *Handler) Start(c telebot.Context) error {
	referralCode := c.Message().Payload
	_, err := h.botService.Login(c , referralCode)
	if err != nil {
		return h.botService.HandleError(c, err)
	}

	// استفاده از HTML برای جلوگیری از خطای کاراکترهای خاص
	msg := fmt.Sprintf(
		"👋 <b>سلام %s عزیز، به پرمیا خوش آمدید!</b> 🌟\n\n"+
			"ما دسترسی شما را به برترین ابزارهای هوش مصنوعی جهان (ChatGPT، Gemini، Claude) با <b>تحویل آنی</b> و <b>قیمت استثنایی</b> فراهم می‌کنیم. 🚀\n\n"+
			"💎 <b>چرا پرمیا؟</b>\n"+
			"✅ تحویل اتوماتیک در کسری از ثانیه\n"+
			"✅ اکانت‌های قانونی و بدون قطعی\n"+
			"✅ پشتیبانی اختصاصی و گارانتی\n\n"+
			"👇 <b>همین الان سرویس مورد نظرتان را انتخاب کنید:</b>",
		escapeHTML(c.Sender().FirstName), // جهت اطمینان، نام کاربر را Escape می‌کنیم
	)

	// Reset user state to none when starting
	h.botService.SetUserState(c.Sender().ID, 0) // StateNone

	return c.Send(msg, &telebot.SendOptions{
		ParseMode:   telebot.ModeHTML, // تغییر به HTML
		ReplyMarkup: menus.MainMenuMarkup,
	})
}

// تابع کمکی برای ایمن‌سازی نام کاربر در حالت HTML
func escapeHTML(s string) string {
	replacer := strings.NewReplacer(
		"<", "&lt;",
		">", "&gt;",
		"&", "&amp;",
	)
	return replacer.Replace(s)
}