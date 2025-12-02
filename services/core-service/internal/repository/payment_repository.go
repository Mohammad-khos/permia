package repository

import (
	"Permia/core-service/internal/domain"
	"context"
	"time"

	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) domain.PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx interface{}, payment *domain.Payment) error {
	return r.db.WithContext(ctx.(context.Context)).Create(payment).Error
}

func (r *paymentRepository) GetByID(ctx interface{}, id uint) (*domain.Payment, error) {
	var payment domain.Payment
	err := r.db.WithContext(ctx.(context.Context)).
		Preload("Order").
		Preload("User").
		First(&payment, "id = ?", id).Error
	return &payment, err
}

func (r *paymentRepository) GetByOrderID(ctx interface{}, orderID uint) (*domain.Payment, error) {
	var payment domain.Payment
	err := r.db.WithContext(ctx.(context.Context)).
		Preload("Order").
		Preload("User").
		First(&payment, "order_id = ?", orderID).Error
	return &payment, err
}

func (r *paymentRepository) UpdateStatus(ctx interface{}, id uint, status domain.PaymentStatus) error {
	return r.db.WithContext(ctx.(context.Context)).
		Model(&domain.Payment{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *paymentRepository) UpdateVerification(ctx interface{}, id uint, status domain.PaymentStatus, verificationURL string) error {
	return r.db.WithContext(ctx.(context.Context)).
		Model(&domain.Payment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":           status,
			"reference_number": verificationURL,
			"verified_at":      time.Now(),
			"updated_at":       time.Now(),
		}).Error
}

func (r *paymentRepository) GetByTransactionID(ctx interface{}, transactionID string) (*domain.Payment, error) {
	var payment domain.Payment
	err := r.db.WithContext(ctx.(context.Context)).
		Preload("Order").
		Preload("User").
		First(&payment, "transaction_id = ?", transactionID).Error
	return &payment, err
}
