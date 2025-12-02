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

	// Register text message handlers for interactive flows
	registerMessageHandlers(bot, menuHandler, sessionRepo, sugar)

	sugar.Info("🤖 Bot is starting...")
	bot.Start()
}

// registerMessageHandlers registers text message handlers for interactive bot flows
func registerMessageHandlers(bot *telebot.Bot, menuHandler *menus.Handler, sessionRepo repository.SessionRepository, logger *zap.SugaredLogger) {
	// Handle text messages
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		text := c.Text()
		userID := c.Sender().ID
		logger.Debugf("Received text from %d: %s", userID, text)

		// Get current user state
		state := sessionRepo.GetState(userID)

		// Handle state-specific inputs first
		if state == domain.StateWaitingForAmount {
			// User is expected to enter an amount
			amount, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
			if err != nil || amount <= 0 {
				return c.Send("❌ مقدار نامعتبر است. لطفا عدد معتبر وارد کنید.")
			}
			// Reset state after processing
			sessionRepo.SetState(userID, domain.StateNone)
			return menuHandler.ProcessChargeAmount(c, text)
		}

		// Handle main menu buttons (these should reset state)
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
			// Set state to waiting for amount
			sessionRepo.SetState(userID, domain.StateWaitingForAmount)
			return menuHandler.ChargeWallet(c)
		}

		// Handle category selection (text version)
		if len(text) > 2 && text[0:2] == "📁 " {
			sessionRepo.SetState(userID, domain.StateNone)
			category := strings.TrimPrefix(text, "📁 ")
			return menuHandler.ShowProducts(c, category)
		}

		// Handle product purchase (text version)
		// Check if it looks like a product selection (contains price indicator)
		if len(text) > 2 && text[len(text)-1] == 'T' {
			sessionRepo.SetState(userID, domain.StateNone)
			// Extract product name and price
			logger.Infof("User %d selecting product: %s", userID, text)
			return menuHandler.ProcessProductOrder(c, text, 0)
		}

		// Default response for unhandled text
		return c.Send("❓ متوجه نشدم. لطفا از دکمه‌های منو استفاده کنید.", &telebot.SendOptions{
			ReplyMarkup: menus.MainMenuMarkup,
		})
	})

	// Handle callback queries (inline buttons)
	bot.Handle(telebot.OnCallback, func(c telebot.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		logger.Debugf("Received callback from %d: %s", userID, data)

		// Acknowledge the callback
		defer c.Respond()

		// Reset state for all callback queries (navigation)
		sessionRepo.SetState(userID, domain.StateNone)

		// Handle main menu callback
		if data == "main_menu" {
			return menuHandler.MainMenu(c)
		}

		// Handle main menu buttons
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

		// Handle wallet charge button
		if data == "charge_wallet" {
			// Set state to waiting for amount
			sessionRepo.SetState(userID, domain.StateWaitingForAmount)
			return menuHandler.ChargeWallet(c)
		}

		// Handle category selection via callback
		if strings.HasPrefix(data, "category:") {
			category := strings.TrimPrefix(data, "category:")
			return menuHandler.ShowProducts(c, category)
		}

		// Handle product selection via callback
		if strings.HasPrefix(data, "product:") {
			// Parse product data
			productData := strings.Split(strings.TrimPrefix(data, "product:"), "|")
			if len(productData) >= 2 {
				productTitle := productData[0]
				price, err := strconv.ParseFloat(productData[1], 64)
				if err != nil {
					return c.Send("❌ خطا در پردازش محصول.")
				}
				return menuHandler.ProcessProductOrder(c, productTitle, price)
			}
		}

		// Default response for unhandled callbacks
		return c.Send("❓ متوجه نشدم. لطفا از دکمه‌های منو استفاده کنید.")
	})
}

// parseAmount is a helper to check if text is a valid amount
func parseAmount(text string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(text), 64)
}