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
	inlineCategoryMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	var inlineCatRows []telebot.Row

	for cat := range categories {
		btn := categoryMarkup.Text(fmt.Sprintf("📁 %s", cat))
		catRows = append(catRows, categoryMarkup.Row(btn))

		inlineBtn := inlineCategoryMarkup.Data(fmt.Sprintf("📁 %s", cat), fmt.Sprintf("category:%s", cat))
		inlineCatRows = append(inlineCatRows, inlineCategoryMarkup.Row(inlineBtn))
	}

	catRows = append(catRows, categoryMarkup.Row(BtnBackToMain))
	inlineCatRows = append(inlineCatRows, inlineCategoryMarkup.Row(inlineCategoryMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")))

	categoryMarkup.Reply(catRows...)
	inlineCategoryMarkup.Inline(inlineCatRows...)

	msg := "🛍️ <b>دسته‌ای را انتخاب کنید:</b>"

	return c.Send(msg, &telebot.SendOptions{
		ParseMode:   telebot.ModeHTML,
		ReplyMarkup: inlineCategoryMarkup,
	})
}

// Profile shows user information (Static Info + Real Balance)
func (h *Handler) Profile(c telebot.Context) error {
	userID := c.Sender().ID
	h.logger.Infof("User %d viewing profile", userID)

	// دریافت موجودی واقعی از Core
	user, err := h.botService.GetProfile(c)
	if err != nil {
		return c.Send("❌ بارگذاری پروفایل ناموفق بود.")
	}

	subs, err := h.botService.GetUserSubscriptions(userID)
	if err != nil {
		h.logger.Errorf("Failed to get subs: %v", err)
	}

	// ✅ تغییرات اینجا اعمال شد:
	// استفاده از دیتای استاتیک (ثابت) اما مفید به جای ID و تاریخ عضویت
	msg := fmt.Sprintf(
		"👤 <b>پروفایل شما</b>\n\n"+
			"🔰 <b>وضعیت حساب:</b> ✅ تایید شده\n"+ // استاتیک
			"⭐️ <b>سطح کاربری:</b> ویژه (VIP)\n"+ // استاتیک (حس خوب به کاربر)
			"💰 <b>موجودی:</b> %.0f تومان\n\n"+ // دینامیک (واقعی)
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

	// Simple pass-through for dates
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

	productsMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	var prodRows []telebot.Row
	inlineProductsMarkup := &telebot.ReplyMarkup{ResizeKeyboard: true}
	var inlineProdRows []telebot.Row

	for _, p := range filtered {
		btn := productsMarkup.Text(fmt.Sprintf("%s - %.0f T", p.Name, p.Price))
		prodRows = append(prodRows, productsMarkup.Row(btn))

		inlineBtn := inlineProductsMarkup.Data(fmt.Sprintf("%s - %.0f T", p.Name, p.Price), fmt.Sprintf("product:%s|%.0f", p.Name, p.Price))
		inlineProdRows = append(inlineProdRows, inlineProductsMarkup.Row(inlineBtn))
	}

	prodRows = append(prodRows, productsMarkup.Row(BtnBackToMain))
	inlineProdRows = append(inlineProdRows, inlineProductsMarkup.Row(inlineProductsMarkup.Data("🔙 بازگشت به منوی اصلی", "main_menu")))

	productsMarkup.Reply(prodRows...)
	inlineProductsMarkup.Inline(inlineProdRows...)

	msg := fmt.Sprintf("📦 <b>%s</b>\n\nبرای خرید یک محصول انتخاب کنید:", h.escapeHTML(category))

	return c.Send(msg, &telebot.SendOptions{
		ParseMode:   telebot.ModeHTML,
		ReplyMarkup: inlineProductsMarkup,
	})
}

func (h *Handler) ProcessProductOrder(c telebot.Context, productTitle string, price float64) error {
	h.logger.Infof("User %d ordering product: %s", c.Sender().ID, productTitle)

	user, err := h.botService.GetProfile(c)
	if err != nil {
		return c.Send("❌ خطا در پردازش سفارش.")
	}

	sku := extractSKU(productTitle)

	order, err := h.coreClient.CreateOrder(user.ID, sku)
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
		return c.Send(fmt.Sprintf("❌ ثبت سفارش ناموفق بود: %v", err))
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

func extractSKU(title string) string {
	parts := strings.Split(title, " - ")
	if len(parts) > 0 {
		return strings.ToLower(strings.ReplaceAll(parts[0], " ", "-"))
	}
	return "unknown"
}