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

	// Register callback query handler
	registerCallbackHandler(bot, menuHandler, sessionRepo, sugar)

	// Register text message handlers
	registerMessageHandlers(bot, menuHandler, sessionRepo, sugar)
	botCommands := []telebot.Command{
		{Text: "start", Description: "🚀 شروع و نمایش منوی اصلی"},
		// اگر کامندهای دیگری دارید می‌توانید اینجا اضافه کنید
		// {Text: "help", Description: "راهنما"},
	}

	if err := bot.SetCommands(botCommands); err != nil {
		sugar.Errorf("Failed to set bot commands: %v", err)
	}
	sugar.Info("🤖 Bot is starting...")
	bot.Start()
}

func registerCallbackHandler(bot *telebot.Bot, menuHandler *menus.Handler, sessionRepo repository.SessionRepository, logger *zap.SugaredLogger) {
	bot.Handle(telebot.OnCallback, func(c telebot.Context) error {
		data := strings.TrimSpace(c.Data())
		userID := c.Sender().ID
		logger.Debugf("Received callback from %d: '%s'", userID, data)

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
		
		// برای محصولاتی که از لیست اینلاین انتخاب می‌شوند
		if strings.HasPrefix(data, "product:") {
			cleanData := strings.TrimPrefix(data, "product:")
			productData := strings.Split(cleanData, "|")
			if len(productData) >= 2 {
				productTitle := productData[0]
				price, err := strconv.ParseFloat(productData[1], 64)
				if err != nil {
					return c.Send("❌ خطا در پردازش قیمت محصول.")
				}
				return menuHandler.ProcessProductOrder(c, productTitle, price)
			}
			return c.Send("❌ اطلاعات محصول نامعتبر است.")
		}

		// Handle subscription details
		if strings.HasPrefix(data, "sub:") {
			idStr := strings.TrimPrefix(data, "sub:")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err == nil {
				return menuHandler.ShowSubscriptionDetail(c, id)
			}
		}

		return c.Send("❓ متوجه نشدم. لطفا از دکمه‌های منو استفاده کنید.")
	})
}

func registerMessageHandlers(bot *telebot.Bot, menuHandler *menus.Handler, sessionRepo repository.SessionRepository, logger *zap.SugaredLogger) {
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		text := strings.TrimSpace(c.Text())
		textLower := strings.ToLower(text)
		userID := c.Sender().ID
		logger.Debugf("Received text from %d: '%s'", userID, text)

		state := sessionRepo.GetState(userID)

		// State-based handlers
		if state == domain.StateWaitingForAmount {
			amount, err := strconv.ParseFloat(text, 64)
			if err != nil || amount <= 0 {
				return c.Send("❌ مقدار نامعتبر است. لطفا عدد معتبر وارد کنید.")
			}
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.ProcessChargeAmount(c, text)
		}

		// Static Menu Buttons
		if text == "🛒 خرید اشتراک" {
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.Buy(c)
		}
		if text == "👤 پروفایل" {
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.Profile(c)
		}
		if text == "💳 کیف پول" {
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.Wallet(c)
		}
		if text == "📞 پشتیبانی" {
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.Support(c)
		}
		if text == "🔙 بازگشت به منوی اصلی" {
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.MainMenu(c)
		}
		if text == "➕ شارژ کیف پول" {
			sessionRepo.SetState(userID, domain.StateWaitingForAmount)
			return menuHandler.ChargeWallet(c)
		}

		// Dynamic Category Handler (New Logic)
		// چک می‌کنیم آیا متن پیام حاوی نام دسته‌ها هست یا خیر
		// چون دکمه‌ها شامل آیکون هستند (مثل "🤖 chatgpt")، باید با contains چک کنیم
		if strings.Contains(textLower, "chatgpt") || 
		   strings.Contains(textLower, "claude") || 
		   strings.Contains(textLower, "gemini") || 
		   strings.Contains(textLower, "tools") || 
		   strings.Contains(text, "📂") { // برای دسته‌های متفرقه که آیکون پوشه دارند
			
			sessionRepo.SetState(userID, domain.StateNone)
			
			// استخراج نام تمیز دسته از متن دکمه
			category := cleanCategoryName(text)
			return menuHandler.ShowProducts(c, category)
		}

		// Fallback for Product Selection (Text Mode) - if needed
		if strings.Contains(text, " - ") && strings.HasSuffix(text, " T") {
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.ProcessProductOrder(c, text, 0)
		}

		return c.Send("❓ متوجه نشدم. لطفا از دکمه‌های منو استفاده کنید.", &telebot.SendOptions{
			ReplyMarkup: menus.MainMenuMarkup,
		})
	})
}

// تابع کمکی برای پاک کردن آیکون از نام دسته
func cleanCategoryName(text string) string {
	// لیست تمام آیکون‌های استفاده شده
	icons := []string{"🤖", "🧠", "✨", "🛠", "📂"}
	clean := text
	for _, icon := range icons {
		clean = strings.ReplaceAll(clean, icon, "")
	}
	return strings.TrimSpace(clean)
}