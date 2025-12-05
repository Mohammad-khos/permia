package service

import (
	"Permia/bot-service/internal/domain"
	"Permia/bot-service/internal/repository"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

// CoreClient defines the interface for the core service client.
type CoreClient interface {
	LoginUser(telegramID int64, username, firstName, lastName , referalCode string) (*domain.User, error)
	GetProfile(telegramID int64) (*domain.User, error)
	GetProducts() ([]domain.Product, error)
	GetUserSubscriptions(telegramID int64) ([]domain.Subscription, error) // Added
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

func (s *BotService) Login(c telebot.Context , referralCode string) (*domain.User, error) {
	user, err := s.coreClient.LoginUser(c.Sender().ID, c.Sender().Username, c.Sender().FirstName, c.Sender().LastName , referralCode)
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
		return s.Login(c , "") // تلاش برای ورود مجدد کاربر
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

// GetUserSubscriptions delegates to core client
func (s *BotService) GetUserSubscriptions(telegramID int64) ([]domain.Subscription, error) {
	subs, err := s.coreClient.GetUserSubscriptions(telegramID)
	if err != nil {
		return nil, err
	}
	return subs, nil
}

func (s *BotService) GetSubscriptions(telegramID int64) ([]domain.Subscription, error) {
	return s.coreClient.GetUserSubscriptions(telegramID)
}

func (s *BotService) SetUserState(userID int64, state domain.UserState) {
	s.sessionRepo.SetState(userID, state)
}

func (s *BotService) GetUserState(userID int64) domain.UserState {
	return s.sessionRepo.GetState(userID)
}

func (s *BotService) GetBotUsername() string {
	return s.bot.Me.Username
}