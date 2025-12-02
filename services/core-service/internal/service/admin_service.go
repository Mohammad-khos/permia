package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"Permia/core-service/internal/domain"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AdminService struct {
	accountRepo domain.AccountRepository
	orderRepo   domain.OrderRepository
	productRepo domain.ProductRepository
	db          *gorm.DB
	logger      *zap.Logger
}

func NewAdminService(
	accountRepo domain.AccountRepository,
	orderRepo domain.OrderRepository,
	productRepo domain.ProductRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *AdminService {
	return &AdminService{
		accountRepo: accountRepo,
		orderRepo:   orderRepo,
		productRepo: productRepo,
		db:          db,
		logger:      logger,
	}
}

// AddInventory adds new accounts to inventory
func (s *AdminService) AddInventory(ctx context.Context, req *domain.AdminInventoryRequest) (map[string]interface{}, error) {
	s.logger.Info("Adding inventory",
		zap.String("product_sku", req.ProductSKU),
		zap.Int("count", req.Count),
	)

	// 1. Check if product exists
	product, err := s.productRepo.GetBySKU(ctx, req.ProductSKU)
	if err != nil {
		s.logger.Warn("Product not found",
			zap.String("product_sku", req.ProductSKU),
			zap.Error(err),
		)
		return nil, errors.New("product not found")
	}

	// 2. Create new accounts
	var accounts []domain.AccountInventory
	for i := 0; i < req.Count; i++ {
		account := domain.AccountInventory{
			ProductSKU:   req.ProductSKU,
			Email:        fmt.Sprintf("%s_%d@domain.com", req.Email, i),
			Password:     req.Password,
			Additional:   req.Additional,
			MaxUsers:     req.MaxUsers,
			CurrentUsers: 0,
			Status:       "AVAILABLE",
			PurchasedAt:  time.Now(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		accounts = append(accounts, account)
	}

	// 3. Save to database
	if err := s.accountRepo.CreateBatch(ctx, accounts); err != nil {
		s.logger.Error("Failed to create batch inventory",
			zap.String("product_sku", req.ProductSKU),
			zap.Int("count", req.Count),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to add accounts: %w", err)
	}

	s.logger.Info("Inventory added successfully",
		zap.String("product_sku", req.ProductSKU),
		zap.Int("count", req.Count),
	)

	return map[string]interface{}{
		"product_sku":   req.ProductSKU,
		"added_count":   req.Count,
		"product_price": product.Price,
		"total_value":   product.Price * float64(req.Count),
		"message":       fmt.Sprintf("%d accounts added successfully", req.Count),
	}, nil
}

// GetInventoryStats gets inventory statistics
func (s *AdminService) GetInventoryStats(ctx context.Context, productSKU string) (*domain.AdminInventoryStats, error) {
	s.logger.Info("Getting inventory stats", zap.String("product_sku", productSKU))

	var totalAccounts, availableAccounts, soldAccounts, expiredAccounts int64
	var availableRevenue float64

	// 1. Count accounts by status
	if err := s.db.WithContext(ctx).
		Model(&domain.AccountInventory{}).
		Where("product_sku = ?", productSKU).
		Count(&totalAccounts).Error; err != nil {
		s.logger.Error("Failed to count total accounts", zap.Error(err))
		return nil, err
	}

	if err := s.db.WithContext(ctx).
		Model(&domain.AccountInventory{}).
		Where("product_sku = ? AND status = ?", productSKU, "AVAILABLE").
		Count(&availableAccounts).Error; err != nil {
		s.logger.Error("Failed to count available accounts", zap.Error(err))
		return nil, err
	}

	if err := s.db.WithContext(ctx).
		Model(&domain.AccountInventory{}).
		Where("product_sku = ? AND status = ?", productSKU, "FILLED").
		Count(&soldAccounts).Error; err != nil {
		s.logger.Error("Failed to count sold accounts", zap.Error(err))
		return nil, err
	}

	if err := s.db.WithContext(ctx).
		Model(&domain.AccountInventory{}).
		Where("product_sku = ? AND status = ?", productSKU, "EXPIRED").
		Count(&expiredAccounts).Error; err != nil {
		s.logger.Error("Failed to count expired accounts", zap.Error(err))
		return nil, err
	}

	// 2. Calculate available revenue
	product, err := s.productRepo.GetBySKU(ctx, productSKU)
	if err != nil {
		s.logger.Error("Failed to get product", zap.String("product_sku", productSKU), zap.Error(err))
		return nil, err
	}

	availableRevenue = float64(availableAccounts) * product.Price

	stats := &domain.AdminInventoryStats{
		TotalAccounts:     int(totalAccounts),
		AvailableAccounts: int(availableAccounts),
		SoldAccounts:      int(soldAccounts),
		ExpiredAccounts:   int(expiredAccounts),
		AvailableRevenue:  availableRevenue,
		ProductSKU:        productSKU,
		Category:          product.Category,
	}

	s.logger.Info("Inventory stats retrieved",
		zap.String("product_sku", productSKU),
		zap.Int("total", stats.TotalAccounts),
		zap.Int("available", stats.AvailableAccounts),
	)

	return stats, nil
}

// GetOrderStats دریافت آمار سفارشات
func (s *AdminService) GetOrderStats(ctx context.Context) (*domain.AdminOrderStats, error) {
	if s.db == nil {
		return nil, errors.New("database connection is not available")
	}

	var totalOrders, pendingOrders, paidOrders, completedOrders, failedOrders int64
	var totalRevenue, pendingRevenue, completedRevenue float64

	// شمارش سفارشات برحسب وضعیت
	s.db.WithContext(ctx).Model(&domain.Order{}).Count(&totalOrders)
	s.db.WithContext(ctx).Model(&domain.Order{}).Where("status = ?", domain.OrderPending).Count(&pendingOrders)
	s.db.WithContext(ctx).Model(&domain.Order{}).Where("status = ?", domain.OrderPaid).Count(&paidOrders)
	s.db.WithContext(ctx).Model(&domain.Order{}).Where("status = ?", domain.OrderCompleted).Count(&completedOrders)
	s.db.WithContext(ctx).Model(&domain.Order{}).Where("status = ?", domain.OrderFailed).Count(&failedOrders)

	// محاسبه درآمد
	s.db.WithContext(ctx).Model(&domain.Order{}).Select("COALESCE(SUM(amount), 0)").Row().Scan(&totalRevenue)
	s.db.WithContext(ctx).Model(&domain.Order{}).Where("status = ?", domain.OrderPending).Select("COALESCE(SUM(amount), 0)").Row().Scan(&pendingRevenue)
	s.db.WithContext(ctx).Model(&domain.Order{}).Where("status = ?", domain.OrderCompleted).Select("COALESCE(SUM(amount), 0)").Row().Scan(&completedRevenue)

	return &domain.AdminOrderStats{
		TotalOrders:      int(totalOrders),
		PendingOrders:    int(pendingOrders),
		PaidOrders:       int(paidOrders),
		CompletedOrders:  int(completedOrders),
		FailedOrders:     int(failedOrders),
		TotalRevenue:     totalRevenue,
		PendingRevenue:   pendingRevenue,
		CompletedRevenue: completedRevenue,
	}, nil
}

// CompleteOrder تکمیل یک سفارش توسط ادمین
func (s *AdminService) CompleteOrder(ctx context.Context, orderID uint) (map[string]interface{}, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, errors.New("سفارش پیدا نشد")
	}

	if order.Status == domain.OrderCompleted {
		return nil, errors.New("سفارش قبلا تکمیل شده است")
	}

	// بروزرسانی وضعیت سفارش
	if err := s.orderRepo.UpdateStatus(ctx, order.ID, string(domain.OrderCompleted)); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"order_id": order.ID,
		"status":   string(domain.OrderCompleted),
		"message":  "سفارش با موفقیت تکمیل شد",
	}, nil
}
