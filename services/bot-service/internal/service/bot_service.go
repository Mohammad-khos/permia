package service

import (
	"Permia/bot-service/internal/domain"
	"Permia/bot-service/internal/repository"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

// CoreClient defines the interface for the core service client.
type CoreClient interface {
	LoginUser(telegramID int64, username, firstName, lastName string) (*domain.User, error)
	GetProfile(telegramID int64) (*domain.User, error)
	GetProducts() ([]domain.Product, error)
}

// BotService handles the core logic of the bot.
type BotService struct {
	bot         *telebot.Bot
	coreClient  CoreClient
	sessionRepo repository.SessionRepository
	logger      *zap.SugaredLogger
}

// NewBotService creates a new BotService.
func NewBotService(
	bot *telebot.Bot,
	coreClient CoreClient,
	sessionRepo repository.SessionRepository,
	logger *zap.SugaredLogger,
) *BotService {
	return &BotService{
		bot:         bot,
		coreClient:  coreClient,
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

func (s *BotService) HandleError(c telebot.Context, err error) error {
	s.logger.Errorf("An error occurred: %v", err)
	// Check for core service unavailability
	if err != nil && strings.Contains(err.Error(), "core service is unavailable") {
		return c.Send("سامانه در حال نگهداری است 🛠")
	}
	return c.Send("❌ خطای غیرمنتظره‌ای رخ داد. لطفا بعدا دوباره تلاش کنید.")
}

func (s *BotService) Login(c telebot.Context) (*domain.User, error) {
	user, err := s.coreClient.LoginUser(c.Sender().ID, c.Sender().Username, c.Sender().FirstName, c.Sender().LastName)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *BotService) GetProfile(c telebot.Context) (*domain.User, error) {
	user, err := s.coreClient.GetProfile(c.Sender().ID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		// This can happen if user was deleted from DB but tries to use the bot
		// We can try to re-login them.
		return s.Login(c)
	}
	return user, nil
}

func (s *BotService) GetProducts() ([]domain.Product, error) {
	products, err := s.coreClient.GetProducts()
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (s *BotService) FormatProducts(products []domain.Product) string {
	if len(products) == 0 {
		return "📭 در حال حاضر محصولی موجود نیست."
	}

	var builder strings.Builder
	builder.WriteString("*🛒 محصولات موجود*\n\n")
	for _, p := range products {
		builder.WriteString(fmt.Sprintf("*%s*\n", escapeMarkdown(p.Name)))
		builder.WriteString(fmt.Sprintf("`%s`\n", escapeMarkdown(p.Description)))
		builder.WriteString(fmt.Sprintf("قیمت: *%.0f تومان*\n\n", p.Price))
	}
	return builder.String()
}

func (s *BotService) FormatProfile(user *domain.User) string {
	var builder strings.Builder
	builder.WriteString("*👤 پروفایل شما*\n\n")
	builder.WriteString(fmt.Sprintf("شناسه تلگرام: `%d`\n", user.TelegramID))
	builder.WriteString(fmt.Sprintf("نام کاربری: @%s\n", escapeMarkdown(user.Username)))
	builder.WriteString(fmt.Sprintf("💰 موجودی فعلی: *%.0f تومان*\n", user.Balance))
	return builder.String()
}

// SetUserState sets the state of a user
func (s *BotService) SetUserState(userID int64, state domain.UserState) {
	s.sessionRepo.SetState(userID, state)
}

// GetUserState gets the state of a user
func (s *BotService) GetUserState(userID int64) domain.UserState {
	return s.sessionRepo.GetState(userID)
}

// escapeMarkdown escapes characters that have special meaning in MarkdownV2.
func escapeMarkdown(s string) string {
	var result strings.Builder
	for _, r := range s {
		if strings.ContainsRune("._*~`>#+-=|{}!", r) {
			result.WriteRune('\\')
		}
		result.WriteRune(r)
	}
	return result.String()
}