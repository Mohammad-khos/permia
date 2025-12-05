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
	// Main Menu (دکمه‌های کیبورد پایین)
	MainMenuMarkup = &telebot.ReplyMarkup{ResizeKeyboard: true}
	BtnBuy         = MainMenuMarkup.Text("🛒 خرید اشتراک")
	BtnProfile     = MainMenuMarkup.Text("👤 پروفایل")
	BtnWallet      = MainMenuMarkup.Text("💳 کیف پول")
	BtnSupport     = MainMenuMarkup.Text("📞 پشتیبانی")
	BtnReferral    = MainMenuMarkup.Text("🔗 دریافت لینک دعوت")

	// Back Button
	BackMarkup    = &telebot.ReplyMarkup{ResizeKeyboard: true}
	BtnBackToMain = BackMarkup.Text("🔙 بازگشت به منوی اصلی")
	
	// Wallet Menu
	WalletMarkup    = &telebot.ReplyMarkup{ResizeKeyboard: true}
	BtnChargeWallet = WalletMarkup.Text("➕ شارژ کیف پول")
	
	// Coupons Button
	BtnMyCoupons = MainMenuMarkup.Text("🎁 کدهای تخفیف من")
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
		MainMenuMarkup.Row(BtnReferral),
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

// MainMenu نمایش منوی اصلی (دکمه‌های پایین)
func (h *Handler) MainMenu(c telebot.Context) error {
	msg := "🏠 **منوی اصلی**\n\nچه کاری می‌خواهید انجام دهید؟"

	return c.Send(msg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: MainMenuMarkup,
	})
}

// Buy Flow
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

	categories := make(map[string]bool)
	for _, p := range products {
		categories[p.Category] = true
	}

	categoryMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	var catRows []telebot.Row

	for cat := range categories {
		emoji := h.getCategoryEmoji(cat)
		btn := categoryMarkup.Text(fmt.Sprintf("%s %s", emoji, cat))
		catRows = append(catRows, categoryMarkup.Row(btn))
	}

	catRows = append(catRows, categoryMarkup.Row(BtnBackToMain))
	categoryMarkup.Reply(catRows...)

	msg := "🛍️ دسته‌ای را انتخاب کنید:"

	return c.Send(msg, &telebot.SendOptions{
		ReplyMarkup: categoryMarkup,
	})
}

// Profile shows user information and subscriptions
func (h *Handler) Profile(c telebot.Context) error {
	h.logger.Infof("User %d viewing profile", c.Sender().ID)

	// ۱. دریافت اطلاعات کاربر
	user, err := h.botService.GetProfile(c)
	if err != nil {
		h.logger.Errorf("Failed to get profile: %v", err)
		return c.Send("❌ بارگذاری پروفایل ناموفق بود. لطفا دوباره تلاش کنید.")
	}

	// ۲. دریافت اشتراک‌های فعال
	subs, err := h.botService.GetSubscriptions(c.Sender().ID)
	// اگر ارور داد مهم نیست، لیست خالی نشان می‌دهیم (نباید کل پروفایل قطع شود)
	if err != nil {
		h.logger.Warnf("Failed to fetch subs for user %d: %v", c.Sender().ID, err)
	}

	safeUsername := h.escapeMarkdown(user.Username)
	if safeUsername == "" {
		safeUsername = "بدون نام کاربری"
	}

	// ۳. ساخت متن پروفایل
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"👤 *پروفایل شما*\n\n"+
			"*نام کاربری:* @%s\n"+
			"*شناسه تلگرام:* `%d`\n"+
			"*موجودی:* %.0f تومان\n"+
			"*تعداد دعوت‌ها:* %d نفر\n\n",
		safeUsername,
		c.Sender().ID,
		user.Balance,
		user.TotalReferrals,
	))

	// ۴. اضافه کردن لیست اشتراک‌ها به پیام
	sb.WriteString("📦 *اشتراک‌های فعال شما:*\n")
	
	if len(subs) == 0 {
		sb.WriteString("_(هیچ سرویس فعالی ندارید)_\n")
	} else {
		for _, sub := range subs {
			// ایمن کردن متن‌ها برای مارک‌داون
			pName := h.escapeMarkdown(sub.ProductName)
			expDate := h.escapeMarkdown(sub.ExpiresAt)
			delData := h.escapeMarkdown(sub.DeliveredData)

			sb.WriteString(fmt.Sprintf(
				"➖➖➖➖➖➖\n"+
				"💎 *%s*\n"+
				"📅 انقضا: %s\n"+
				"🔑 اطلاعات:\n`%s`\n",
				pName, expDate, delData,
			))
		}
	}
	profileMenu := &telebot.ReplyMarkup{ResizeKeyboard: true}
    profileMenu.Reply(
        profileMenu.Row(BtnMyCoupons), // اضافه کردن دکمه کوپن
        profileMenu.Row(BtnBackToMain),
    )
	// دکمه بازگشت
	inlineBackMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnBack := inlineBackMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")
	inlineBackMarkup.Inline(inlineBackMarkup.Row(btnBack))

	return c.Send(sb.String(), &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: profileMenu,
	})
}
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

func (h *Handler) Support(c telebot.Context) error {
	supportMsg := "📞 *پشتیبانی*\n\n" +
		"برای هرگونه مشکل یا سوال، با ما تماس بگیرید:\n\n" +
		"📧 ایمیل: support@permia\\.com\n" +
		"💬 تلگرام: @AdminID\n\n" +
		"ما آماده کمک هستیم\\!"

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
		return c.Send("❌ بارگذاری محصولات ناموفق بود.")
	}

	var filtered []domain.Product
	for _, p := range products {
		if p.Category == category {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) == 0 {
		return c.Send("📭 در این دسته محصولی موجود نیست.")
	}

	inlineProductsMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	var inlineProdRows []telebot.Row

	for _, p := range filtered {
		// استفاده از Title اگر موجود بود، وگرنه نام ساده
		displayName := p.Title
		if displayName == "" {
			displayName = fmt.Sprintf("محصول %.0f", p.Price)
		}
		
		buttonText := fmt.Sprintf("%s - %.0f T", displayName, p.Price)
		
		inlineBtn := inlineProductsMarkup.Data(
			buttonText,
			fmt.Sprintf("product:%s", p.SKU),
		)
		inlineProdRows = append(inlineProdRows, inlineProductsMarkup.Row(inlineBtn))
	}

	inlineProdRows = append(inlineProdRows, inlineProductsMarkup.Row(inlineProductsMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")))

	inlineProductsMarkup.Inline(inlineProdRows...)

	emoji := h.getCategoryEmoji(category)
	msg := fmt.Sprintf("%s *%s*\n\nبرای خرید یک محصول انتخاب کنید:", emoji, h.escapeMarkdown(category))

	return c.Send(msg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: inlineProductsMarkup,
	})
}

// PreviewInvoice (پیش‌فاکتور)
func (h *Handler) PreviewInvoice(c telebot.Context, sku string) error {
	products, err := h.botService.GetProducts()
	if err != nil {
		return c.Send("❌ خطا در دریافت اطلاعات محصول.")
	}

	var targetProduct domain.Product
	found := false
	for _, p := range products {
		if p.SKU == sku {
			targetProduct = p
			found = true
			break
		}
	}

	if !found {
		return c.Send("❌ محصول مورد نظر یافت نشد.")
	}

	description := targetProduct.Description
	if description == "" {
		description = "توضیحات در دسترس نیست."
	}

	invoiceMsg := fmt.Sprintf(
		"🧾 *پیش‌فاکتور سفارش*\n\n"+
			"🛍 *محصول:* %s\n"+
			"📝 *توضیحات:* %s\n"+
			"💰 *مبلغ قابل پرداخت:* %.0f تومان\n\n"+
			"⚠️ لطفا قبل از تایید نهایی، اطلاعات بالا را بررسی کنید\\.\n",
		h.escapeMarkdown(targetProduct.Title),
		h.escapeMarkdown(description),
		targetProduct.Price,
	)

	confirmMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	// ارسال pay:SKU برای تایید نهایی
	btnConfirm := confirmMarkup.Data("✅ تایید و پرداخت نهایی", fmt.Sprintf("pay:%s", sku))
	btnCoupon := confirmMarkup.Data("🎟 ثبت کد تخفیف", fmt.Sprintf("coupon:%s", sku)) // دکمه جدید
	btnCancel := confirmMarkup.Data("❌ انصراف", "main_menu")

	confirmMarkup.Inline(
		confirmMarkup.Row(btnConfirm),
		confirmMarkup.Row(btnCoupon), // اضافه شد
		confirmMarkup.Row(btnCancel),
	)

	err = c.EditOrSend(invoiceMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: confirmMarkup,
	})

	if err != nil {
		h.logger.Errorf("Failed to send invoice message: %v", err)
		msg := fmt.Sprintf("🧾 پیش‌فاکتور سفارش\n\n🛍 محصول: %s\n📝 توضیحات: %s\n💰 مبلغ: %.0f تومان\n\n⚠️ لطفا بررسی و تایید کنید.",
			targetProduct.Title, description, targetProduct.Price)
		return c.EditOrSend(msg, &telebot.SendOptions{
			ReplyMarkup: confirmMarkup,
		})
	}
	return nil
}

// ProcessProductOrder (خرید نهایی - اصلاح شده با ۳ آرگومان)
func (h *Handler) ProcessProductOrder(c telebot.Context, sku string) error {
	h.logger.Infof("User %d ordering sku: %s", c.Sender().ID, sku)

	user, err := h.botService.GetProfile(c)
	if err != nil {
		return c.Send("❌ خطا در پردازش سفارش.")
	}

	// ✅ اصلاح شده: فراخوانی CreateOrder فقط با ۳ آرگومان (مطابق client.go)
	couponCode := h.botService.GetDraft(c.Sender().ID, "active_coupon")
	order, err := h.coreClient.CreateOrder(user.ID, c.Sender().ID, sku ,couponCode)

	// پاک‌کردن کوپن و پیش‌نویس پس از استفاده
	h.botService.ClearDraft(c.Sender().ID)
	
	if err != nil {
		h.logger.Errorf("Failed to create order: %v", err)
		if strings.Contains(err.Error(), "insufficient") {
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
		return c.Send("❌ ثبت سفارش ناموفق بود.")
	}

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
	h.botService.SetUserState(c.Sender().ID, domain.StateWaitingForAmount)
	return c.Send("💰 *مقدار شارژ را \\(به تومان\\) وارد کنید:*\n\nمثال: 100000", &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdownV2,
	})
}

// ProcessChargeAmount handles the amount input
func (h *Handler) ProcessChargeAmount(c telebot.Context, amountStr string) error {
	amount, err := strconv.ParseFloat(strings.TrimSpace(amountStr), 64)
	if err != nil || amount <= 0 {
		return c.Send("❌ مقدار نامعتبر است. لطفا عدد معتبر وارد کنید.")
	}

	h.logger.Infof("User %d charging wallet with amount: %.0f", c.Sender().ID, amount)

	user, err := h.botService.GetProfile(c)
	if err != nil {
		h.logger.Errorf("Failed to get user for payment: %v", err)
		return c.Send("❌ خطا در پردازش پرداخت. لطفا دوباره تلاش کنید.")
	}

	paymentLink, err := h.coreClient.GetPaymentLink(user.ID, amount)
	if err != nil {
		h.logger.Errorf("Failed to get payment link: %v", err)
		return c.Send("❌ ایجاد لینک پرداخت ناموفق بود. لطفا دوباره تلاش کنید.")
	}

	inlineMarkup := &telebot.ReplyMarkup{}
	btn := inlineMarkup.URL("💳 پرداخت با زرین‌پال", paymentLink)
	
	backBtn := inlineMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")
	
	inlineMarkup.Inline(
		inlineMarkup.Row(btn),
		inlineMarkup.Row(backBtn),
	)

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

func (h *Handler) GetReferralLink(c telebot.Context) error {
	user, err := h.botService.GetProfile(c)
	if err != nil {
		return c.Send("❌ خطا در دریافت اطلاعات کاربری.")
	}

	botUsername := h.botService.GetBotUsername() 
	refLink := fmt.Sprintf("https://t.me/%s?start=%s", botUsername, user.ReferralCode)

	msg := fmt.Sprintf(
		"🎁 **دعوت از دوستان**\n\n"+
			"با دعوت دوستان خود به پرمیا، در خریدهای آن‌ها شریک شوید!\n\n"+
			"🔗 **لینک اختصاصی شما:**\n`%s`\n\n"+
			"👥 **تعداد دعوت‌های شما:** %d نفر\n\n"+
			"👇 لینک بالا را برای دوستانتان ارسال کنید.",
		refLink,
		user.TotalReferrals,
	)

	return c.Send(msg, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// GetMyCoupons نمایش کدهای تخفیف کاربر
func (h *Handler) GetMyCoupons(c telebot.Context) error {
    coupons, err := h.coreClient.GetUserCoupons(c.Sender().ID)
    if err != nil || len(coupons) == 0 {
        return c.Send("📭 شما در حال حاضر کد تخفیف فعالی ندارید.")
    }

    var sb strings.Builder
    sb.WriteString("🎁 **کدهای تخفیف شما:**\n\n")
    for _, coup := range coupons {
        sb.WriteString(fmt.Sprintf("🎟 کد: `%s`\n٪ تخفیف: %.0f%%\n\n", coup.Code, coup.Percent))
    }
    
    return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// وقتی کاربر روی دکمه "ثبت کد تخفیف" زد
func (h *Handler) AskForCoupon(c telebot.Context, sku string) error {
    h.botService.SetDraft(c.Sender().ID, "sku_for_coupon", sku)
    h.botService.SetUserState(c.Sender().ID, domain.StateWaitingForCoupon)
    return c.Send("🎟 لطفا کد تخفیف خود را ارسال کنید:")
}

// وقتی کاربر کد را نوشت
// نسخه اصلاح شده تابع ValidateAndApplyCoupon
func (h *Handler) ValidateAndApplyCoupon(c telebot.Context, code string) error {
	userID := c.Sender().ID
	sku := h.botService.GetDraft(userID, "sku_for_coupon")
	
	products, _ := h.botService.GetProducts()
	var price float64
	var title string
	for _, p := range products {
		if p.SKU == sku {
			price = p.Price
			title = p.Title
			break
		}
	}

	newPrice, err := h.coreClient.ValidateCoupon(userID, code, price)
	if err != nil {
		h.botService.SetUserState(userID, domain.StateNone)
		return c.Send(fmt.Sprintf("❌ خطا: %v", err))
	}

	h.botService.SetDraft(userID, "active_coupon", code)
	h.botService.SetUserState(userID, domain.StateNone)

	// ✅ اصلاح شده: اضافه کردن \\ قبل از !
	msg := fmt.Sprintf(
		"✅ *کد تخفیف اعمال شد\\!*\n\n🛍 محصول: %s\n💰 قیمت جدید: %.0f T",
		h.escapeMarkdown(title), newPrice,
	)

	confirmMarkup := &telebot.ReplyMarkup{}
	btnPay := confirmMarkup.Data("✅ پرداخت مبلغ نهایی", fmt.Sprintf("pay:%s", sku))
	confirmMarkup.Inline(confirmMarkup.Row(btnPay))

	return c.Send(msg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: confirmMarkup,
	})
}

// Helper function to escape markdown characters
func (h *Handler) escapeMarkdown(s string) string {
	var result strings.Builder
	specialChars := "_*[]()~`>#+-=|{}.!"
	for _, r := range s {
		if strings.ContainsRune(specialChars, r) {
			result.WriteRune('\\')
		}
		result.WriteRune(r)
	}
	return result.String()
}

// Helper: اختصاص ایموجی به دسته‌بندی‌ها
func (h *Handler) getCategoryEmoji(category string) string {
	catLower := strings.ToLower(category)
	if strings.Contains(catLower, "gpt") {
		return "🤖"
	}
	if strings.Contains(catLower, "gemini") {
		return "💎"
	}
	if strings.Contains(catLower, "claude") {
		return "🎭"
	}
	if strings.Contains(catLower, "midjourney") || strings.Contains(catLower, "art") {
		return "🎨"
	}
	if strings.Contains(catLower, "tool") {
		return "🔧"
	}
	return "📂"
}