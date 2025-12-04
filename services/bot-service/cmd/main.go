package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"Permia/bot-service/internal/config"
	"Permia/bot-service/internal/domain"
	"Permia/bot-service/internal/handler"
	"Permia/bot-service/internal/handler/commands"
	"Permia/bot-service/internal/handler/menus"
	"Permia/bot-service/internal/infrastructure/core"
	"Permia/bot-service/internal/repository"
	"Permia/bot-service/internal/service"

	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

func main() {
	// Initialize Logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// Load Config
	cfg, err := config.LoadConfig("../../../deployment/.env")
	if err != nil {
		sugar.Fatalf("Failed to load configuration: %v", err)
	}

	sugar.Infof("Bot configuration loaded - Token: %s..., Core API: %s",
		cfg.TelegramBotToken[:10], cfg.CoreApiURL)

	// Initialize Telebot
	pref := telebot.Settings{
		Token:  cfg.TelegramBotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}
	bot, err := telebot.NewBot(pref)
	if err != nil {
		sugar.Fatalf("Failed to create bot: %v", err)
	}

	// Initialize Core Service Client
	coreClient := core.NewClient(cfg.CoreApiURL, sugar)

	// Initialize Session Repository
	sessionRepo := repository.NewInMemorySessionRepository()

	// Initialize Bot Service
	botSvc := service.NewBotService(bot, coreClient, sessionRepo, sugar)

	// Initialize Handlers
	commandHandler := commands.NewHandler(botSvc)
	menuHandler := menus.NewHandler(botSvc, coreClient, sugar)

	// Register all handlers
	h := handler.New(bot, commandHandler, menuHandler)
	h.Register()

	// Register callback query handler BEFORE starting bot
	registerCallbackHandler(bot, menuHandler, sessionRepo, sugar)

	// Register text message handlers for interactive flows
	registerMessageHandlers(bot, menuHandler, sessionRepo, sugar)

	// ---------------------------------------------------------
	// تنظیمات منوی بات و توضیحات (Bot Profile)
	// ---------------------------------------------------------
	
	// 1. تنظیم منوی دستورات (کنار کادر متن)
	botCommands := []telebot.Command{
		{Text: "start", Description: "🚀 شروع و نمایش منوی اصلی"},
	}
	if err := bot.SetCommands(botCommands); err != nil {
		sugar.Errorf("Failed to set bot commands: %v", err)
	}

	// 2. تنظیم متن خوش‌آمدگویی (صفحه خالی قبل از استارت)
	welcomeDesc := "👋 سلام همراه گرامی، به دنیای هوش مصنوعی پرمیا خوش آمدید! 🌟\n\n" +
		"ما در Permia دسترسی شما را به قدرتمندترین ابزارهای AI جهان (مثل ChatGPT، Gemini و Claude) با کمترین هزینه و بالاترین سرعت ممکن می‌کنیم. 🚀\n\n" +
		"💎 چرا پرمیا انتخاب حرفه‌ای‌هاست؟\n" +
		"✅ تحویل آنی: بلافاصله پس از پرداخت، اکانت خود را دریافت کنید.\n" +
		"✅ قیمت رقابتی: حذف واسطه‌ها و ارائه بهترین قیمت بازار.\n" +
		"✅ تضمین کیفیت: تمام اکانت‌ها قانونی و با گارانتی کامل هستند.\n" +
		"✅ پشتیبانی: تیم ما همیشه در کنار شماست.\n\n" +
		"👇 برای شروع روی دکمه Start کلیک کنید"

	// اصلاح شده: فراخوانی مستقیم تابع با دو آرگومان (متن توضیحات، زبان)
	// آرگومان دوم خالی ("") به معنی زبان پیش‌فرض است
	if err := bot.SetMyDescription(welcomeDesc, ""); err != nil {
		sugar.Warnf("Failed to set bot description: %v", err)
	}
	// ---------------------------------------------------------

	sugar.Info("🤖 Bot is starting...")
	bot.Start()
}

// registerCallbackHandler registers callback query handlers (inline buttons)
func registerCallbackHandler(bot *telebot.Bot, menuHandler *menus.Handler, sessionRepo repository.SessionRepository, logger *zap.SugaredLogger) {
	bot.Handle(telebot.OnCallback, func(c telebot.Context) error {
		data := c.Callback().Data
		if strings.HasPrefix(data, "\f") {
			data = strings.TrimPrefix(data, "\f")
		}

		userID := c.Sender().ID
		logger.Debugf("Received callback from %d: %s", userID, data)

		defer c.Respond()

		sessionRepo.SetState(userID, domain.StateNone)

		if data == "main_menu" {
			return menuHandler.MainMenu(c)
		}

		if data == "buy" {
			return menuHandler.Buy(c)
		}
		if data == "profile" {
			return menuHandler.Profile(c)
		}
		if data == "wallet" {
			return menuHandler.Wallet(c)
		}
		if data == "support" {
			return menuHandler.Support(c)
		}

		if data == "charge_wallet" {
			sessionRepo.SetState(userID, domain.StateWaitingForAmount)
			return menuHandler.ChargeWallet(c)
		}

		if strings.HasPrefix(data, "category:") {
			category := strings.TrimPrefix(data, "category:")
			return menuHandler.ShowProducts(c, category)
		}

		if strings.HasPrefix(data, "product:") {
			sku := strings.TrimPrefix(data, "product:")
			if sku != "" {
				return menuHandler.PreviewInvoice(c, sku)
			}
		}

		if strings.HasPrefix(data, "pay:") {
			sku := strings.TrimPrefix(data, "pay:")
			if sku != "" {
				return menuHandler.ProcessProductOrder(c, sku)
			}
		}

		return c.Send("❓ متوجه نشدم. لطفا از دکمه‌های منو استفاده کنید.")
	})
}

// registerMessageHandlers registers text message handlers for interactive bot flows
func registerMessageHandlers(bot *telebot.Bot, menuHandler *menus.Handler, sessionRepo repository.SessionRepository, logger *zap.SugaredLogger) {
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		text := c.Text()
		userID := c.Sender().ID
		logger.Debugf("Received text from %d: %s", userID, text)

		state := sessionRepo.GetState(userID)

		if state == domain.StateWaitingForAmount {
			amount, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
			if err != nil || amount <= 0 {
				return c.Send("❌ مقدار نامعتبر است. لطفا عدد معتبر وارد کنید.")
			}
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.ProcessChargeAmount(c, text)
		}

		// Handle Main Menu Actions
		switch text {
		case "🛒 خرید اشتراک":
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.Buy(c)
		case "👤 پروفایل":
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.Profile(c)
		case "💳 کیف پول":
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.Wallet(c)
		case "📞 پشتیبانی":
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.Support(c)
		case "🔙 بازگشت به منوی اصلی":
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.MainMenu(c)
		case "➕ شارژ کیف پول":
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.ChargeWallet(c)
		}

		// Handle Category Selection (Dynamic Emojis)
		if isCategory, catName := extractCategory(text); isCategory {
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.ShowProducts(c, catName)
		}

		return c.Send("❓ متوجه نشدم. لطفا از دکمه‌های منو استفاده کنید.", &telebot.SendOptions{
			ReplyMarkup: menus.MainMenuMarkup,
		})
	})
}

// extractCategory checks if the text starts with a known category emoji prefix
func extractCategory(text string) (bool, string) {
	// لیست ایموجی‌هایی که در menus.go استفاده می‌شوند
	prefixes := []string{"📂 ", "🤖 ", "💎 ", "🎭 ", "🎨 ", "🚀 ", "🔧 "}
	
	for _, p := range prefixes {
		if strings.HasPrefix(text, p) {
			return true, strings.TrimPrefix(text, p)
		}
	}
	return false, ""
}