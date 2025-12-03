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
	msg := "🏠 منوی اصلی\n\nچه کاری می‌خواهید انجام دهید؟"

	// Create inline markup for main menu
	inlineMainMenuMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnBuy := inlineMainMenuMarkup.Data("🛒 خرید اشتراک", "buy")
	btnProfile := inlineMainMenuMarkup.Data("👤 پروفایل", "profile")
	btnWallet := inlineMainMenuMarkup.Data("💳 کیف پول", "wallet")
	btnSupport := inlineMainMenuMarkup.Data("📞 پشتیبانی", "support")

	inlineMainMenuMarkup.Inline(
		inlineMainMenuMarkup.Row(btnBuy, btnProfile),
		inlineMainMenuMarkup.Row(btnWallet, btnSupport),
	)

	return c.Send(msg, &telebot.SendOptions{
		ReplyMarkup: inlineMainMenuMarkup,
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

	// Create category buttons (both text and inline versions)
	// Text buttons for backward compatibility
	categoryMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	var catRows []telebot.Row

	// Also create inline buttons with callback data
	inlineCategoryMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	var inlineCatRows []telebot.Row

	for cat := range categories {
		// Text button
		btn := categoryMarkup.Text(fmt.Sprintf("📁 %s", cat))
		catRows = append(catRows, categoryMarkup.Row(btn))

		// Inline button with callback
		inlineBtn := inlineCategoryMarkup.Data(fmt.Sprintf("📁 %s", h.escapeMarkdown(cat)), fmt.Sprintf("category:%s", cat))
		inlineCatRows = append(inlineCatRows, inlineCategoryMarkup.Row(inlineBtn))
	}

	// Add back button to both
	catRows = append(catRows, categoryMarkup.Row(BtnBackToMain))
	inlineCatRows = append(inlineCatRows, inlineCategoryMarkup.Row(inlineCategoryMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")))

	// Set up markups
	categoryMarkup.Reply(catRows...)
	inlineCategoryMarkup.Inline(inlineCatRows...)

	msg := "🛍️ دسته‌ای را انتخاب کنید:"

	// Send with inline markup for better UX
	return c.Send(msg, &telebot.SendOptions{
		ReplyMarkup: inlineCategoryMarkup,
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
			"*عضویت از:* به‌زودی",
		user.Username,
		c.Sender().ID,
	)

	// Create inline markup for back button
	inlineBackMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnBack := inlineBackMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")
	inlineBackMarkup.Inline(inlineBackMarkup.Row(btnBack))

	return c.Send(profileMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: inlineBackMarkup,
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

	// Create inline markup for charge button
	inlineWalletMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnCharge := inlineWalletMarkup.Data("➕ شارژ کیف پول", "charge_wallet")
	btnBack := inlineWalletMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")
	inlineWalletMarkup.Inline(
		inlineWalletMarkup.Row(btnCharge),
		inlineWalletMarkup.Row(btnBack),
	)

	return c.Send(walletMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: inlineWalletMarkup,
	})
}

// Support shows support information
func (h *Handler) Support(c telebot.Context) error {
	supportMsg := "📞 *پشتیبانی*\n\n" +
		"برای هرگونه مشکل یا سوال، با ما تماس بگیرید:\n\n" +
		"📧 ایمیل: support@permia\\.com\n" +
		"💬 تلگرام: @AdminID\n\n" +
		"ما آماده کمک هستیم\\!"

	// Create inline markup for back button
	inlineBackMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnBack := inlineBackMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")
	inlineBackMarkup.Inline(inlineBackMarkup.Row(btnBack))

	return c.Send(supportMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: inlineBackMarkup,
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
		if p.Category == category {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) == 0 {
		return c.Send("📭 در این دسته محصولی موجود نیست.")
	}

	// Create product selection buttons (both text and inline versions)
	// Text buttons for backward compatibility
	productsMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	var prodRows []telebot.Row

	// Also create inline buttons with callback data
	inlineProductsMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	var inlineProdRows []telebot.Row

	for _, p := range filtered {
		// Text button
		btn := productsMarkup.Text(fmt.Sprintf("%s - %.0f T", p.Name, p.Price))
		prodRows = append(prodRows, productsMarkup.Row(btn))

		// Inline button with callback
		displayName := h.escapeMarkdown(p.Name)
		inlineBtn := inlineProductsMarkup.Data(fmt.Sprintf("%s - %.0f T", displayName, p.Price), fmt.Sprintf("product:%s|%.0f", p.Name, p.Price))
		inlineProdRows = append(inlineProdRows, inlineProductsMarkup.Row(inlineBtn))
	}

	// Add back button to both
	prodRows = append(prodRows, productsMarkup.Row(BtnBackToMain))
	inlineProdRows = append(inlineProdRows, inlineProductsMarkup.Row(inlineProductsMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")))

	// Set up markups
	productsMarkup.Reply(prodRows...)
	inlineProductsMarkup.Inline(inlineProdRows...)

	msg := fmt.Sprintf("📦 *%s*\n\nبرای خرید یک محصول انتخاب کنید:",
		h.escapeMarkdown(category))

	// Send with inline markup for better UX
	return c.Send(msg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: inlineProductsMarkup,
	})
}

// ProcessProductOrder handles product selection and creates order
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
			// Create inline markup for wallet button
			insufficientMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
			btnCharge := insufficientMarkup.Data("➕ شارژ کیف پول", "charge_wallet")
			btnBack := insufficientMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")
			insufficientMarkup.Inline(
				insufficientMarkup.Row(btnCharge),
				insufficientMarkup.Row(btnBack),
			)

			return c.Send("💸 *موجودی ناکافی*\n\nکیف پول شما به اندازه کافی شارژ ندارد\\. \n\nآیا می‌خواهید کیف پول خود را شارژ کنید؟", &telebot.SendOptions{
				ParseMode:   telebot.ModeMarkdownV2,
				ReplyMarkup: insufficientMarkup,
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

	// Create inline markup for main menu button
	successMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnMainMenu := successMarkup.Data("🏠 منوی اصلی", "main_menu")
	successMarkup.Inline(successMarkup.Row(btnMainMenu))

	return c.Send(deliveryMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: successMarkup,
	})

}

// ChargeWallet initiates wallet charging
func (h *Handler) ChargeWallet(c telebot.Context) error {
	h.logger.Infof("User %d requesting wallet charge", c.Sender().ID)

	// Set user state to waiting for amount
	h.botService.SetUserState(c.Sender().ID, domain.StateWaitingForAmount)

	return c.Send("💰 *مقدار شارژ را \\(به تومان\\) وارد کنید:*\n\nمثال: 100000", &telebot.SendOptions{
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

	// Get user to get user ID
	user, err := h.botService.GetProfile(c)
	if err != nil {
		h.logger.Errorf("Failed to get user for payment: %v", err)
		return c.Send("❌ خطا در پردازش پرداخت. لطفا دوباره تلاش کنید.")
	}

	// Get payment link from core service
	paymentLink, err := h.coreClient.GetPaymentLink(user.ID, amount)
	if err != nil {
		h.logger.Errorf("Failed to get payment link: %v", err)
		return c.Send("❌ ایجاد لینک پرداخت ناموفق بود. لطفا دوباره تلاش کنید.")
	}

	// Create inline keyboard with payment link
	inlineMarkup := &telebot.ReplyMarkup{}
	btn := inlineMarkup.URL("💳 پرداخت با زرین‌پال", paymentLink)
	inlineMarkup.Inline(inlineMarkup.Row(btn))

	chargeMsg := fmt.Sprintf(
		"💰 *شارژ کیف پول*\n\n"+
			"*مبلغ:* %.0f تومان\n\n"+
			"برای تکمیل پرداخت دکمه زیر را بزنید\\.",
		amount,
	)

	// Create back button
	backBtn := inlineMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")
	inlineMarkup.Inline(
		inlineMarkup.Row(btn),
		inlineMarkup.Row(backBtn),
	)

	return c.Send(chargeMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: inlineMarkup,
	})
}

// Helper function to escape markdown characters
func (h *Handler) escapeMarkdown(s string) string {
	var result strings.Builder
	for _, r := range s {
		if strings.ContainsRune("._*~`>#+-=|{}!", r) {
			result.WriteRune('\\')
		}
		result.WriteRune(r)
	}
	return result.String()
}

// Helper function to extract SKU from product title
func extractSKU(title string) string {
	// Remove price part and extract SKU
	parts := strings.Split(title, " - ")
	if len(parts) > 0 {
		// Return the first part as SKU (product name)
		return strings.ToLower(strings.ReplaceAll(parts[0], " ", "-"))
	}
	return "unknown"
}
