package commands

import (
	"Permia/bot-service/internal/handler/menus"
	"Permia/bot-service/internal/service"

	"gopkg.in/telebot.v3"
)

type Handler struct {
	botService *service.BotService
}

func NewHandler(botService *service.BotService) *Handler {
	return &Handler{botService: botService}
}

// Start handles the /start command.
func (h *Handler) Start(c telebot.Context) error {
	_, err := h.botService.Login(c)
	if err != nil {
		return h.botService.HandleError(c, err)
	}

	// Send welcome message
	welcomeMsg := "Welcome to *Permia* Bot\\! Your one\\-stop shop for AI accounts\\."

	// Reset user state to none when starting
	h.botService.SetUserState(c.Sender().ID, 0) // StateNone

	return c.Send(welcomeMsg, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: menus.MainMenuMarkup,
	})
}