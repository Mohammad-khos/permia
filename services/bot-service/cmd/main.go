package main

import (
	"log"
	"time"

	"Permia/bot-service/internal/config"
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

	// Register text message handlers for interactive flows
	registerMessageHandlers(bot, menuHandler, sugar)

	sugar.Info("🤖 Bot is starting...")
	bot.Start()
}

// registerMessageHandlers registers text message handlers for interactive bot flows
func registerMessageHandlers(bot *telebot.Bot, menuHandler *menus.Handler, logger *zap.SugaredLogger) {
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		text := c.Text()
		logger.Debugf("Received text from %d: %s", c.Sender().ID, text)

		// Handle main menu buttons
		if text == "🛒 خرید اشتراک" {
			return menuHandler.Buy(c)
		}
		if text == "👤 پروفایل" {
			return menuHandler.Profile(c)
		}
		if text == "💳 کیف پول" {
			return menuHandler.Wallet(c)
		}
		if text == "📞 پشتیبانی" {
			return menuHandler.Support(c)
		}
		if text == "🔙 بازگشت به منوی اصلی" {
			return menuHandler.MainMenu(c)
		}
		if text == "➕ شارژ کیف پول" {
			return menuHandler.ChargeWallet(c)
		}

		// Handle category selection
		if len(text) > 2 && text[0:2] == "📁 " {
			return menuHandler.ShowProducts(c, text)
		}

		// Handle product purchase
		// Check if it looks like a product selection (contains price indicator)
		if len(text) > 2 && text[len(text)-1] == 'T' {
			// Extract product name and price
			// This is simplified - in production you'd store state
			logger.Infof("User %d selecting product: %s", c.Sender().ID, text)
			return menuHandler.ProcessProductOrder(c, text, 0)
		}

		// Handle wallet charge amount input
		// If we're expecting a number (wallet charge flow)
		if _, err := parseAmount(text); err == nil {
			return menuHandler.ProcessChargeAmount(c, text)
		}

		// Default response for unhandled text
		return c.Send("❓ متوجه نشدم. لطفا از دکمه\u200cهای منو استفاده کنید.", &telebot.SendOptions{
			ReplyMarkup: menus.MainMenuMarkup,
		})
	})
}

// parseAmount is a helper to check if text is a valid amount
func parseAmount(text string) (float64, error) {
	// This is called by the message handler above
	// Will be used to parse wallet charge amounts
	var amount float64
	_ = text // Placeholder to avoid unused import warning
	return amount, nil
}
