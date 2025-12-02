package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"Permia/core-service/internal/domain"

	"gorm.io/gorm"
)

type OrderService struct {
	orderRepo   domain.OrderRepository
	userRepo    domain.UserRepository
	productRepo domain.ProductRepository
	accountRepo domain.AccountRepository
	vccProvider domain.VCCProvider
	db          *gorm.DB
}

func NewOrderService(
	orderRepo domain.OrderRepository,
	userRepo domain.UserRepository,
	productRepo domain.ProductRepository,
	accountRepo domain.AccountRepository,
	vccProvider domain.VCCProvider,
	db *gorm.DB,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		userRepo:    userRepo,
		productRepo: productRepo,
		accountRepo: accountRepo,
		vccProvider: vccProvider,
		db:          db,
	}
}

type PurchaseResult struct {
	OrderID       uint
	Status        string
	DeliveredData string
}

func (s *OrderService) PurchaseFlow(ctx context.Context, userID uint, productSKU string) (*PurchaseResult, error) {
	var result PurchaseResult

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. دریافت محصول و چک موجودی (کد قبلی را نگه دارید)...
		product, err := s.productRepo.GetBySKU(ctx, productSKU)
		if err != nil {
			return err
		}
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return err
		}
		if user.WalletBalance < product.Price {
			return errors.New("موجودی کافی نیست")
		}

		// 2. منطق تخصیص (اینجا تغییر می‌کند)
		var assignedAccount *domain.AccountInventory
		var deliveryInfo string
		orderStatus := domain.OrderPending

		// اگر محصول نیاز به صدور کارت دارد (مثلاً "اختصاصی قانونی" یا "لینک دعوت")
		switch product.Type {
		case "private_legal", "private_invite":
			// --- [بخش جدید: استفاده از Brocard] ---
			// درخواست کارت: مبلغ 2 دلار (برای وریفای) - نوع کارت "visa_universal"
			card, err := s.vccProvider.IssueCard(2.0, "visa_universal")
			if err != nil {
				return fmt.Errorf("خطا در صدور کارت Brocard: %v", err)
			}

			orderStatus = domain.OrderCompleted
			deliveryInfo = fmt.Sprintf(
				"Card Info for Activation:\nPAN: %s\nCVV: %s\nExp: %s\n\n*Use US IP Address*",
				card.PAN, card.CVV, card.Expiry,
			)
			// -------------------------------------

		case "manual_order":
			orderStatus = domain.OrderPaid
			deliveryInfo = "سفارش ثبت شد. منتظر انجام توسط پشتیبانی."

		default:
			// محصولات انبار (مثل اشتراکی)
			account, err := s.accountRepo.GetAvailableAccount(ctx, product.SKU)
			if err != nil {
				return errors.New("موجودی انبار تمام شده")
			}
			assignedAccount = account
			if err := s.accountRepo.MarkAsSold(ctx, account.ID); err != nil {
				return err
			}

			orderStatus = domain.OrderCompleted
			deliveryInfo = fmt.Sprintf("Email: %s\nData: %s", account.Email, account.Additional)
		}

		// 3. کسر پول و ثبت سفارش (کد قبلی)...
		if err := s.userRepo.UpdateWallet(ctx, userID, -product.Price); err != nil {
			return err
		}

		newOrder := domain.Order{
			OrderNumber:   fmt.Sprintf("ORD-%d-%d", userID, time.Now().Unix()),
			UserID:        userID,
			ProductID:     product.ID,
			Amount:        product.Price,
			Status:        orderStatus,
			DeliveredData: deliveryInfo,
			CreatedAt:     time.Now(),
		}
		if assignedAccount != nil {
			newOrder.AccountID = &assignedAccount.ID
		}

		if err := s.orderRepo.Create(ctx, &newOrder); err != nil {
			return err
		}

		result.OrderID = newOrder.ID
		result.Status = string(orderStatus)
		result.DeliveredData = deliveryInfo
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAllOrders بازیابی تمام سفارشات
func (s *OrderService) GetAllOrders(ctx context.Context) ([]domain.Order, error) {
	return s.orderRepo.GetAllOrders(ctx)
}

// GetOrderByID بازیابی سفارش به وسیله ID
func (s *OrderService) GetOrderByID(ctx context.Context, id uint) (*domain.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}
