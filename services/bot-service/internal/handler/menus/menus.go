package menus

import (
	"fmt"
	"strconv"
	"strings"

	"Permia/bot-service/internal/domain"
	"Permia/bot-service/internal/infrastructure/core"
	"Permia/bot-service/internal/service"

	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

var (
	// Main Menu
	MainMenuMarkup = &telebot.ReplyMarkup{ResizeKeyboard: true}
	BtnBuy         = MainMenuMarkup.Text("🛒 خرید اشتراک")
	BtnProfile     = MainMenuMarkup.Text("👤 پروفایل")
	BtnWallet      = MainMenuMarkup.Text("💳 کیف پول")
	BtnSupport     = MainMenuMarkup.Text("📞 پشتیبانی")

	// Back Button
	BackMarkup    = &telebot.ReplyMarkup{ResizeKeyboard: true}
	BtnBackToMain = BackMarkup.Text("🔙 بازگشت به منوی اصلی")

	// Category Buttons (will be set dynamically)
	CategoryMarkup = &telebot.ReplyMarkup{ResizeKeyboard: true}

	// Products by Category (will be set dynamically)
	ProductsMarkup = &telebot.ReplyMarkup{ResizeKeyboard: true}

	// Wallet Menu
	WalletMarkup    = &telebot.ReplyMarkup{ResizeKeyboard: true}
	BtnChargeWallet = WalletMarkup.Text("➕ شارژ کیف پول")
)

type Handler struct {
	botService *service.BotService
	coreClient *core.Client
	logger     *zap.SugaredLogger
}

func NewHandler(botService *service.BotService, coreClient *core.Client, logger *zap.SugaredLogger) *Handler {
	MainMenuMarkup.Reply(
		MainMenuMarkup.Row(BtnBuy, BtnProfile),
		MainMenuMarkup.Row(BtnWallet, BtnSupport),
	)
	BackMarkup.Reply(BackMarkup.Row(BtnBackToMain))
	WalletMarkup.Reply(
		WalletMarkup.Row(BtnChargeWallet),
		WalletMarkup.Row(BtnBackToMain),
	)
	return &Handler{
		botService: botService,
		coreClient: coreClient,
		logger:     logger,
	}
}

func (h *Handler) MainMenu(c telebot.Context) error {
	msg := "🏠 منوی اصلی\n\nچه کاری می\u200cخواهید انجام دهید؟"
	return c.Send(msg, &telebot.SendOptions{
		ReplyMarkup: MainMenuMarkup,
	})
}

// Buy Flow - Shows categories first
func (h *Handler) Buy(c telebot.Context) error {
	h.logger.Infof("User %d viewing buy menu", c.Sender().ID)

	products, err := h.botService.GetProducts()
	if err != nil {
		h.logger.Errorf("Failed to get products: %v", err)
		return c.Send("❌ بارگذاری محصولات ناموفق بود. لطفا دوباره تلاش کنید.")
	}

	if len(products) == 0 {
		return c.Send("📭 در حال حاضر محصولی موجود نیست.")
	}

	// Extract unique categories
	categories := make(map[string]bool)
	for _, p := range products {
		categories[p.Category] = true
	}

	// Create category buttons
	categoryMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	var catRows []telebot.Row

	for cat := range categories {
		btn := categoryMarkup.Text(fmt.Sprintf("📁 %s", cat))
		catRows = append(catRows, categoryMarkup.Row(btn))
	}
	catRows = append(catRows, categoryMarkup.Row(BtnBackToMain))
	categoryMarkup.Reply(catRows...)

	msg := "🛍️ دسته\u200cای را انتخاب کنید:"
	return c.Send(msg, &telebot.SendOptions{
		ReplyMarkup: categoryMarkup,
	})
}

// Profile shows user information
func (h *Handler) Profile(c telebot.Context) error {
	h.logger.Infof("User %d viewing profile", c.Sender().ID)

	user, err := h.botService.GetProfile(c)
	if err != nil {
		h.logger.Errorf("Failed to get profile: %v", err)
		return c.Send("❌ بارگذاری پروفایل ناموفق بود. لطفا دوباره تلاش کنید.")
	}

	profileMsg := fmt.Sprintf(
		"👤 *پروفایل شما*\n\n"+
			"*نام کاربری:* @%s\n"+
			"*شناسه تلگرام:* `%d`\n"+
			"*عضویت از:* به\u200cزودی",
		user.Username,
		c.Sender().ID,
	)

	return c.Send(profileMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: BackMarkup,
	})
}

// Wallet shows balance and charge option
func (h *Handler) Wallet(c telebot.Context) error {
	h.logger.Infof("User %d viewing wallet", c.Sender().ID)

	user, err := h.botService.GetProfile(c)
	if err != nil {
		h.logger.Errorf("Failed to get wallet balance: %v", err)
		return c.Send("❌ بارگذاری کیف پول ناموفق بود. لطفا دوباره تلاش کنید.")
	}

	walletMsg := fmt.Sprintf(
		"💳 *کیف پول شما*\n\n"+
			"*مانده حساب:* %.0f تومان\n\n"+
			"برای شارژ کیف پول دکمه زیر را فشار دهید\\.",
		user.Balance,
	)

	return c.Send(walletMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: WalletMarkup,
	})
}

// Support shows support information
func (h *Handler) Support(c telebot.Context) error {
	supportMsg := "📞 *پشتیبانی*\n\n" +
		"برای هرگونه مشکل یا سوال، با ما تماس بگیرید:\n\n" +
		"📧 ایمیل: support@permia.com\n" +
		"💬 تلگرام: @permia_support\n\n" +
		"ما آماده کمک هستیم\\!"

	return c.Send(supportMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: BackMarkup,
	})
}

// ShowProducts shows products in a selected category
func (h *Handler) ShowProducts(c telebot.Context, category string) error {
	products, err := h.botService.GetProducts()
	if err != nil {
		return c.Send("❌ بارگذاری محصولات ناموفق بود. لطفا دوباره تلاش کنید.")
	}

	// Filter by category
	var filtered []domain.Product
	for _, p := range products {
		if strings.Contains(p.Category, strings.TrimPrefix(category, "📁 ")) {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) == 0 {
		return c.Send("📭 در این دسته محصولی موجود نیست.")
	}

	// Create product selection buttons
	productsMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	var prodRows []telebot.Row

	for _, p := range filtered {
		btn := productsMarkup.Text(fmt.Sprintf("%s - %.0f T", p.Name, p.Price))
		prodRows = append(prodRows, productsMarkup.Row(btn))
	}
	prodRows = append(prodRows, productsMarkup.Row(BtnBackToMain))
	productsMarkup.Reply(prodRows...)

	msg := fmt.Sprintf("📦 *%s*\n\nبرای خرید یک محصول انتخاب کنید:",
		strings.TrimPrefix(category, "📁 "))
	return c.Send(msg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: productsMarkup,
	})
} // ProcessProductOrder handles product selection and creates order
func (h *Handler) ProcessProductOrder(c telebot.Context, productTitle string, price float64) error {
	h.logger.Infof("User %d ordering product: %s", c.Sender().ID, productTitle)

	// Get user first
	user, err := h.botService.GetProfile(c)
	if err != nil {
		return c.Send("❌ خطا در پردازش سفارش.")
	}

	// Extract SKU from product title (simplified)
	sku := extractSKU(productTitle)

	// Create order via core service
	order, err := h.coreClient.CreateOrder(user.ID, sku)
	if err != nil {
		h.logger.Errorf("Failed to create order: %v", err)

		// Check for insufficient funds
		if strings.Contains(err.Error(), "insufficient") {
			return c.Send("💸 *موجودی ناکافی*\n\nکیف پول شما به اندازه کافی شارژ ندارد\\. \n\nآیا می\u200cخواهید کیف پول خود را شارژ کنید؟", &telebot.SendOptions{
				ParseMode:   telebot.ModeMarkdownV2,
				ReplyMarkup: WalletMarkup,
			})
		}

		return c.Send(fmt.Sprintf("❌ ثبت سفارش ناموفق بود: %v", err))
	}

	// Order successful - show delivery data
	deliveryMsg := fmt.Sprintf(
		"✅ *سفارش با موفقیت ثبت شد\\!*\n\n"+
			"*شناسه سفارش:* `%d`\n"+
			"*مبلغ:* %.0f تومان\n\n"+
			"*اطلاعات حساب شما:*\n"+
			"```\n%s\n```",
		order.OrderID,
		order.Amount,
		order.DeliveredData,
	)

	return c.Send(deliveryMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: MainMenuMarkup,
	})
}

// ChargeWallet initiates wallet charging
func (h *Handler) ChargeWallet(c telebot.Context) error {
	h.logger.Infof("User %d requesting wallet charge", c.Sender().ID)
	return c.Send("💰 *مقدار شارژ را (به تومان) وارد کنید:*\n\nمثال: 100000", &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdownV2,
	})
}

// ProcessChargeAmount handles the amount input and creates payment link
func (h *Handler) ProcessChargeAmount(c telebot.Context, amountStr string) error {
	amount, err := strconv.ParseFloat(strings.TrimSpace(amountStr), 64)
	if err != nil || amount <= 0 {
		return c.Send("❌ مقدار نامعتبر است. لطفا عدد معتبر وارد کنید.")
	}

	h.logger.Infof("User %d charging wallet with amount: %.0f", c.Sender().ID, amount)

	// Get payment link from core service
	paymentLink, err := h.coreClient.GetPaymentLink(0, amount) // userID will be set by core service
	if err != nil {
		h.logger.Errorf("Failed to get payment link: %v", err)
		return c.Send("❌ ایجاد لینک پرداخت ناموفق بود. لطفا دوباره تلاش کنید.")
	}

	// Create inline keyboard with payment link
	inlineMarkup := &telebot.ReplyMarkup{}
	btn := inlineMarkup.URL("💳 پرداخت با زرین\u200cپال", paymentLink)
	inlineMarkup.Inline(inlineMarkup.Row(btn))

	chargeMsg := fmt.Sprintf(
		"💰 *شارژ کیف پول*\n\n"+
			"*مبلغ:* %.0f تومان\n\n"+
			"برای تکمیل پرداخت دکمه زیر را بزنید\\.",
		amount,
	)

	return c.Send(chargeMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: inlineMarkup,
	})
}

// Helper function to extract SKU from product title
func extractSKU(title string) string {
	// This is a simplified extraction - in production you'd want better parsing
	parts := strings.Fields(title)
	if len(parts) > 0 {
		return strings.ToLower(parts[0])
	}
	return "unknown"
}
