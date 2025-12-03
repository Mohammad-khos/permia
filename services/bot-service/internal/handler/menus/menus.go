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
	msg := "🏠 <b>منوی اصلی</b>\n\nچه کاری می‌خواهید انجام دهید؟"

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
		ParseMode:   telebot.ModeHTML,
		ReplyMarkup: inlineMainMenuMarkup,
	})
}

func (h *Handler) Buy(c telebot.Context) error {
	h.logger.Infof("User %d viewing buy menu", c.Sender().ID)

	products, err := h.botService.GetProducts()
	if err != nil {
		h.logger.Errorf("Failed to get products: %v", err)
		return c.Send("❌ بارگذاری محصولات ناموفق بود.")
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
	var tempRow []telebot.Btn

	for cat := range categories {
		icon := getCategoryIcon(cat)
		btn := categoryMarkup.Text(fmt.Sprintf("%s %s", icon, cat))
		tempRow = append(tempRow, btn)

		if len(tempRow) == 2 {
			catRows = append(catRows, categoryMarkup.Row(tempRow...))
			tempRow = []telebot.Btn{}
		}
	}
	if len(tempRow) > 0 {
		catRows = append(catRows, categoryMarkup.Row(tempRow...))
	}

	catRows = append(catRows, categoryMarkup.Row(BtnBackToMain))
	categoryMarkup.Reply(catRows...)

	msg := "🛍️ <b>دسته‌ای را انتخاب کنید:</b>"

	return c.Send(msg, &telebot.SendOptions{
		ParseMode:   telebot.ModeHTML,
		ReplyMarkup: categoryMarkup,
	})
}

func (h *Handler) ShowProducts(c telebot.Context, category string) error {
	products, err := h.botService.GetProducts()
	if err != nil {
		return c.Send("❌ بارگذاری محصولات ناموفق بود.")
	}

	var filtered []domain.Product
	for _, p := range products {
		if strings.EqualFold(p.Category, category) {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) == 0 {
		return c.Send("📭 در این دسته محصولی موجود نیست.")
	}

	var msgBuilder strings.Builder
	icon := getCategoryIcon(category)
	msgBuilder.WriteString(fmt.Sprintf("📦 <b>محصولات دسته %s %s:</b>\n\n", icon, h.escapeHTML(category)))

	inlineProductsMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	var inlineProdRows []telebot.Row

	for i, p := range filtered {
		displayName := p.Name
		if displayName == "" {
			displayName = p.SKU
		}
		if displayName == "" {
			displayName = fmt.Sprintf("محصول %d", p.ID)
		}

		msgBuilder.WriteString(fmt.Sprintf("%d. <b>%s</b>\n💰 قیمت: %.0f تومان\n\n", i+1, h.escapeHTML(displayName), p.Price))

		btnText := fmt.Sprintf("%s | %.0f T", displayName, p.Price)
		
		// ✅ تغییر حیاتی: استفاده از p.SKU به جای p.Name
		// این خط باعث می‌شود شناسه درست به سرور ارسال شود
		callbackData := fmt.Sprintf("product:%s|%.0f", p.SKU, p.Price)
		
		inlineBtn := inlineProductsMarkup.Data(btnText, callbackData)
		inlineProdRows = append(inlineProdRows, inlineProductsMarkup.Row(inlineBtn))
	}

	msgBuilder.WriteString("👇 <b>لطفا جهت خرید، یکی از گزینه‌های زیر را انتخاب کنید:</b>")

	btnBack := inlineProductsMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")
	inlineProdRows = append(inlineProdRows, inlineProductsMarkup.Row(btnBack))

	inlineProductsMarkup.Inline(inlineProdRows...)

	return c.Send(msgBuilder.String(), &telebot.SendOptions{
		ParseMode:   telebot.ModeHTML,
		ReplyMarkup: inlineProductsMarkup,
	})
}

// ProcessProductOrder handles order creation
func (h *Handler) ProcessProductOrder(c telebot.Context, productSKU string, price float64) error {
	telegramID := c.Sender().ID
	h.logger.Infof("User %d ordering SKU: %s", telegramID, productSKU)

	user, err := h.botService.GetProfile(c)
	var dbUserID uint = 0
	if err == nil && user != nil {
		dbUserID = user.ID
	}

	// ✅ تغییر حیاتی: حذف extractSKU
	// چون دکمه‌ها الان SKU واقعی را می‌فرستند، نیازی به تبدیل نام به SKU نیست.
	sku := productSKU

	order, err := h.coreClient.CreateOrder(dbUserID, telegramID, sku)
	
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

			return c.Send("💸 <b>موجودی ناکافی</b>\n\nکیف پول شما به اندازه کافی شارژ ندارد.\n\nآیا می‌خواهید کیف پول خود را شارژ کنید؟", &telebot.SendOptions{
				ParseMode:   telebot.ModeHTML,
				ReplyMarkup: insufficientMarkup,
			})
		}

		// نمایش خطای عمومی
		return c.Send("❌ ثبت سفارش ناموفق بود. لطفا با پشتیبانی تماس بگیرید.")
	}

	deliveryMsg := fmt.Sprintf(
		"✅ <b>سفارش با موفقیت ثبت شد!</b>\n\n"+
			"🔢 <b>شناسه سفارش:</b> <code>%d</code>\n"+
			"💰 <b>مبلغ:</b> %.0f تومان\n\n"+
			"🔑 <b>اطلاعات اکانت شما:</b>\n"+
			"<pre>%s</pre>",
		order.OrderID,
		order.Amount,
		h.escapeHTML(order.DeliveredData),
	)

	successMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnMainMenu := successMarkup.Data("🏠 منوی اصلی", "main_menu")
	successMarkup.Inline(successMarkup.Row(btnMainMenu))

	return c.Send(deliveryMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeHTML,
		ReplyMarkup: successMarkup,
	})
}

// ... (بقیه توابع بدون تغییر)

func (h *Handler) Profile(c telebot.Context) error {
	userID := c.Sender().ID
	h.logger.Infof("User %d viewing profile", userID)

	user, err := h.botService.GetProfile(c)
	if err != nil {
		return c.Send("❌ بارگذاری پروفایل ناموفق بود.")
	}

	subs, err := h.botService.GetUserSubscriptions(userID)
	if err != nil {
		h.logger.Errorf("Failed to get subs: %v", err)
	}

	msg := fmt.Sprintf(
		"👤 <b>پروفایل شما</b>\n\n"+
			"🔰 <b>وضعیت حساب:</b> ✅ تایید شده\n"+
			"⭐️ <b>سطح کاربری:</b> ویژه (VIP)\n"+
			"💰 <b>موجودی:</b> %.0f تومان\n\n"+
			"👇 <b>لیست اشتراک‌های شما:</b>",
		user.Balance,
	)

	markup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	var rows []telebot.Row

	if len(subs) > 0 {
		for _, s := range subs {
			btnText := fmt.Sprintf("🟢 %s", s.ProductName)
			btn := markup.Data(btnText, fmt.Sprintf("sub:%d", s.ID))
			rows = append(rows, markup.Row(btn))
		}
	} else {
		msg += "\n\n<i>(هیچ اشتراک فعالی ندارید)</i>"
	}

	rows = append(rows, markup.Row(markup.Data("🔙 بازگشت به منوی اصلی", "main_menu")))
	markup.Inline(rows...)

	return c.Send(msg, &telebot.SendOptions{
		ParseMode:   telebot.ModeHTML,
		ReplyMarkup: markup,
	})
}

func (h *Handler) ShowSubscriptionDetail(c telebot.Context, subID int64) error {
	subs, err := h.botService.GetUserSubscriptions(c.Sender().ID)
	if err != nil {
		return c.Send("❌ خطا در دریافت اطلاعات.")
	}

	var sub *domain.Subscription
	for _, s := range subs {
		if s.ID == subID {
			sub = &s
			break
		}
	}

	if sub == nil {
		return c.Send("❌ اشتراک یافت نشد.")
	}

	convertToJalali := func(d string) string { return d }

	detailMsg := fmt.Sprintf(
		"🎫 <b>اطلاعات اشتراک سرویس</b>\n\n"+
			"📦 <b>سرویس:</b> %s\n"+
			"🔖 <b>شناسه سفارش:</b> <code>%d</code>\n\n"+
			"🔐 <b>اطلاعات اتصال:</b>\n<pre>%s</pre>\n\n"+
			"📅 <b>تاریخ شروع:</b> %s\n"+
			"📅 <b>تاریخ اتمام (میلادی):</b> %s\n"+
			"📅 <b>تاریخ اتمام (شمسی):</b> %s\n\n"+
			"⚠️ <i>لطفا اطلاعات بالا را در دستگاه خود ذخیره کنید.</i>",
		h.escapeHTML(sub.ProductName),
		sub.ID,
		h.escapeHTML(sub.DeliveredData),
		h.escapeHTML(convertToJalali(sub.CreatedAt)),
		h.escapeHTML(sub.ExpiresAt),
		h.escapeHTML(convertToJalali(sub.ExpiresAt)),
	)

	markup := &telebot.ReplyMarkup{}
	btnBack := markup.Data("🔙 بازگشت به لیست", "profile")
	markup.Inline(markup.Row(btnBack))

	return c.Edit(detailMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeHTML,
		ReplyMarkup: markup,
	})
}

func (h *Handler) Wallet(c telebot.Context) error {
	user, err := h.botService.GetProfile(c)
	if err != nil {
		return c.Send("❌ بارگذاری کیف پول ناموفق بود.")
	}

	walletMsg := fmt.Sprintf(
		"💳 <b>کیف پول شما</b>\n\n"+
			"💵 <b>مانده حساب:</b> %.0f تومان\n\n"+
			"برای شارژ کیف پول دکمه زیر را فشار دهید.",
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
		ParseMode:   telebot.ModeHTML,
		ReplyMarkup: inlineWalletMarkup,
	})
}

func (h *Handler) Support(c telebot.Context) error {
	supportMsg := "📞 <b>پشتیبانی</b>\n\n" +
		"برای هرگونه مشکل یا سوال، با ما تماس بگیرید:\n\n" +
		"📧 ایمیل: support@permia.com\n" +
		"💬 تلگرام: @AdminID\n\n" +
		"ما آماده کمک هستیم!"

	inlineBackMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnBack := inlineBackMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")
	inlineBackMarkup.Inline(inlineBackMarkup.Row(btnBack))

	return c.Send(supportMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeHTML,
		ReplyMarkup: inlineBackMarkup,
	})
}

func (h *Handler) ChargeWallet(c telebot.Context) error {
	h.botService.SetUserState(c.Sender().ID, domain.StateWaitingForAmount)
	return c.Send("💰 <b>مقدار شارژ را (به تومان) وارد کنید:</b>\n\nمثال: 100000", &telebot.SendOptions{
		ParseMode: telebot.ModeHTML,
	})
}

func (h *Handler) ProcessChargeAmount(c telebot.Context, amountStr string) error {
	amount, err := strconv.ParseFloat(strings.TrimSpace(amountStr), 64)
	if err != nil || amount <= 0 {
		return c.Send("❌ مقدار نامعتبر است. لطفا عدد معتبر وارد کنید.")
	}

	user, err := h.botService.GetProfile(c)
	if err != nil {
		return c.Send("❌ خطا در پردازش پرداخت.")
	}

	paymentLink, err := h.coreClient.GetPaymentLink(user.ID, amount)
	if err != nil {
		return c.Send("❌ ایجاد لینک پرداخت ناموفق بود.")
	}

	inlineMarkup := &telebot.ReplyMarkup{}
	btn := inlineMarkup.URL("💳 پرداخت با زرین‌پال", paymentLink)
	backBtn := inlineMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")
	inlineMarkup.Inline(
		inlineMarkup.Row(btn),
		inlineMarkup.Row(backBtn),
	)

	chargeMsg := fmt.Sprintf(
		"💰 <b>شارژ کیف پول</b>\n\n"+
			"💳 <b>مبلغ:</b> %.0f تومان\n\n"+
			"برای تکمیل پرداخت دکمه زیر را بزنید.",
		amount,
	)

	return c.Send(chargeMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeHTML,
		ReplyMarkup: inlineMarkup,
	})
}

func (h *Handler) escapeHTML(s string) string {
	replacer := strings.NewReplacer(
		"<", "&lt;",
		">", "&gt;",
		"&", "&amp;",
	)
	return replacer.Replace(s)
}

func getCategoryIcon(cat string) string {
	catLower := strings.ToLower(cat)
	if strings.Contains(catLower, "chatgpt") || strings.Contains(catLower, "gpt") {
		return "🤖"
	}
	if strings.Contains(catLower, "claude") {
		return "🧠"
	}
	if strings.Contains(catLower, "gemini") {
		return "✨"
	}
	if strings.Contains(catLower, "tools") || strings.Contains(catLower, "ابزار") {
		return "🛠"
	}
	return "📁"
}