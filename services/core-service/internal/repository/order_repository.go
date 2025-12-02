package repository

import (
	"Permia/core-service/internal/domain"
	"context"

	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) domain.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) UpdateStatus(ctx context.Context, orderID uint, status string) error {
	return r.db.WithContext(ctx).Model(&domain.Order{}).
		Where("id = ?", orderID).
		Update("status", status).Error
}

func (r *orderRepository) GetHistoryByUserID(ctx context.Context, userID uint) ([]domain.Order, error) {
	var orders []domain.Order
	// Preload("Product") یعنی اطلاعات محصول را هم همراه سفارش بیار (Join)
	err := r.db.WithContext(ctx).
		Preload("Product").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetByID(ctx context.Context, id uint) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("User").
		Preload("Account").
		First(&order, "id = ?", id).Error
	return &order, err
}

func (r *orderRepository) GetAllOrders(ctx context.Context) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("User").
		Preload("Account").
		Order("created_at desc").
		Find(&orders).Error
	return orders, err
}
