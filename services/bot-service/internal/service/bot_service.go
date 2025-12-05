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
	LoginUser(telegramID int64, username, firstName, lastName, referralCode string) (*domain.User, error)
	GetProfile(telegramID int64) (*domain.User, error)
	GetProducts() ([]domain.Product, error)
	GetUserSubscriptions(telegramID int64) ([]domain.Subscription, error)
	ValidateCoupon(telegramID int64, code string, amount float64) (float64, error)
	GetUserCoupons(telegramID int64) ([]domain.Coupon, error)
}

type BotService struct {
	bot         *telebot.Bot
	coreClient  CoreClient
	sessionRepo repository.SessionRepository
	logger      *zap.SugaredLogger
}

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
	if err != nil && strings.Contains(err.Error(), "core service is unavailable") {
		return c.Send("سامانه در حال نگهداری است 🛠")
	}
	return c.Send("❌ خطای غیرمنتظره‌ای رخ داد.")
}

func (s *BotService) Login(c telebot.Context, referralCode string) (*domain.User, error) {
	user, err := s.coreClient.LoginUser(
		c.Sender().ID,
		c.Sender().Username,
		c.Sender().FirstName,
		c.Sender().LastName,
		referralCode,
	)
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
		return s.Login(c, "")
	}
	return user, nil
}

func (s *BotService) GetProducts() ([]domain.Product, error) {
	return s.coreClient.GetProducts()
}

func (s *BotService) GetSubscriptions(telegramID int64) ([]domain.Subscription, error) {
	return s.coreClient.GetUserSubscriptions(telegramID)
}

// متد جدید برای رفع خطا
func (s *BotService) GetBotUsername() string {
	return s.bot.Me.Username
}

// مدیریت وضعیت
func (s *BotService) SetUserState(userID int64, state domain.UserState) {
	s.sessionRepo.SetState(userID, state)
}

func (s *BotService) GetUserState(userID int64) domain.UserState {
	return s.sessionRepo.GetState(userID)
}

// مدیریت Draft (دیتای موقت)
func (s *BotService) SetDraft(userID int64, key, value string) {
	s.sessionRepo.SetDraft(userID, key, value)
}

func (s *BotService) GetDraft(userID int64, key string) string {
	return s.sessionRepo.GetDraft(userID, key)
}

func (s *BotService) ClearDraft(userID int64) {
	s.sessionRepo.ClearDraft(userID)
}