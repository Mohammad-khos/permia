package handler

import (
	"gopkg.in/telebot.v3"
)

// CommandHandler defines the interface for command handlers.
type CommandHandler interface {
	Start(c telebot.Context) error
}

// MenuHandler defines the interface for menu handlers.
type MenuHandler interface {
	MainMenu(c telebot.Context) error
	Buy(c telebot.Context) error
	Profile(c telebot.Context) error
	Wallet(c telebot.Context) error
	Support(c telebot.Context) error
	ShowProducts(c telebot.Context, category string) error
	ProcessProductOrder(c telebot.Context, productTitle string, price float64) error
	ChargeWallet(c telebot.Context) error
	ProcessChargeAmount(c telebot.Context, amountStr string) error
}

type Handler struct {
	bot         *telebot.Bot
	commands    CommandHandler
	menuHandler MenuHandler
}

func New(bot *telebot.Bot, commands CommandHandler, menus MenuHandler) *Handler {
	return &Handler{bot: bot, commands: commands, menuHandler: menus}
}

func (h *Handler) Register() {
	h.bot.Handle("/start", h.commands.Start)
	// OnCallback and OnText handlers are registered in main.go
	// to avoid conflicts and centralize message routing logic
}