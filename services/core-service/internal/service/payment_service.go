package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"Permia/core-service/internal/domain"
	"Permia/core-service/internal/infrastructure/zarinpal"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PaymentService struct {
	paymentRepo    domain.PaymentRepository
	orderRepo      domain.OrderRepository
	userRepo       domain.UserRepository
	zarinpalClient *zarinpal.ZarinpalClient
	db             *gorm.DB
	logger         *zap.Logger
}

func NewPaymentService(
	paymentRepo domain.PaymentRepository,
	orderRepo domain.OrderRepository,
	userRepo domain.UserRepository,
	zarinpalClient *zarinpal.ZarinpalClient,
	db *gorm.DB,
	logger *zap.Logger,
) *PaymentService {
	return &PaymentService{
		paymentRepo:    paymentRepo,
		orderRepo:      orderRepo,
		userRepo:       userRepo,
		zarinpalClient: zarinpalClient,
		db:             db,
		logger:         logger,
	}
}

type ChargeRequest struct {
	OrderID       uint   `json:"order_id" binding:"required"`
	UserID        uint   `json:"user_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"` // card, wallet
}

type ChargeResponse struct {
	PaymentID       uint    `json:"payment_id"`
	OrderID         uint    `json:"order_id"`
	Amount          float64 `json:"amount"`
	Status          string  `json:"status"`
	VerificationURL string  `json:"verification_url"`
	TransactionID   string  `json:"transaction_id"`
}

// Charge initiates payment process
func (s *PaymentService) Charge(ctx context.Context, req *ChargeRequest) (*ChargeResponse, error) {
	s.logger.Info("Initiating payment charge",
		zap.Uint("user_id", req.UserID),
		zap.Uint("order_id", req.OrderID),
		zap.String("payment_method", req.PaymentMethod),
	)

	// 1. Get order
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Order not found", zap.Uint("order_id", req.OrderID))
			return nil, errors.New("order not found")
		}
		s.logger.Error("Failed to get order", zap.Uint("order_id", req.OrderID), zap.Error(err))
		return nil, err
	}

	// 2. Check wallet balance if wallet payment method
	if req.PaymentMethod == "wallet" {
		user, err := s.userRepo.GetByID(ctx, req.UserID)
		if err != nil {
			s.logger.Error("Failed to get user", zap.Uint("user_id", req.UserID), zap.Error(err))
			return nil, err
		}
		if user.WalletBalance < order.Amount {
			s.logger.Warn("Insufficient wallet balance",
				zap.Uint("user_id", req.UserID),
				zap.Float64("balance", user.WalletBalance),
				zap.Float64("amount", order.Amount),
			)
			return nil, errors.New("insufficient wallet balance")
		}
	}

	// 3. Create payment record
	payment := &domain.Payment{
		OrderID:       order.ID,
		UserID:        req.UserID,
		Amount:        order.Amount,
		Status:        domain.PaymentPending,
		PaymentMethod: req.PaymentMethod,
		TransactionID: fmt.Sprintf("TXN-%d-%d", req.UserID, time.Now().Unix()),
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		s.logger.Error("Failed to create payment record", zap.Error(err))
		return nil, err
	}

	verificationURL := ""

	// 4. Handle payment based on method
	switch req.PaymentMethod {
	case "card":
		// Call Zarinpal to request payment
		callbackURL := fmt.Sprintf("https://yourdomain.com/api/v1/payment/verify?payment_id=%d", payment.ID)
		zarinpalResp, err := s.zarinpalClient.RequestPayment(order.Amount, fmt.Sprintf("Order #%d", order.ID), callbackURL)
		if err != nil {
			s.logger.Error("Failed to request payment from Zarinpal", zap.Error(err))
			payment.Status = domain.PaymentFailed
			payment.ErrorMessage = err.Error()
			s.paymentRepo.UpdateStatus(ctx, payment.ID, domain.PaymentFailed)
			return nil, fmt.Errorf("payment gateway error: %w", err)
		}

		payment.Status = domain.PaymentVerifying
		payment.VerificationURL = zarinpal.GetPaymentURL(zarinpalResp.Data.Authority)
		verificationURL = payment.VerificationURL

		if err := s.paymentRepo.UpdateVerification(ctx, payment.ID, domain.PaymentVerifying, zarinpalResp.Data.Authority); err != nil {
			s.logger.Error("Failed to update payment with authority", zap.Error(err))
			return nil, err
		}

	case "wallet":
		// Direct deduction from wallet
		if err := s.userRepo.UpdateWallet(ctx, req.UserID, -order.Amount); err != nil {
			s.logger.Error("Failed to update wallet", zap.Uint("user_id", req.UserID), zap.Error(err))
			return nil, err
		}

		// Update order status
		if err := s.orderRepo.UpdateStatus(ctx, order.ID, string(domain.OrderPaid)); err != nil {
			s.logger.Error("Failed to update order status", zap.Uint("order_id", order.ID), zap.Error(err))
			return nil, err
		}

		now := time.Now()
		payment.Status = domain.PaymentCompleted
		payment.VerifiedAt = &now

		if err := s.paymentRepo.UpdateStatus(ctx, payment.ID, domain.PaymentCompleted); err != nil {
			s.logger.Error("Failed to update payment status", zap.Error(err))
			return nil, err
		}

		s.logger.Info("Wallet payment completed successfully",
			zap.Uint("user_id", req.UserID),
			zap.Float64("amount", order.Amount),
		)
	}

	return &ChargeResponse{
		PaymentID:       payment.ID,
		OrderID:         order.ID,
		Amount:          payment.Amount,
		Status:          string(payment.Status),
		VerificationURL: verificationURL,
		TransactionID:   payment.TransactionID,
	}, nil
}

type VerifyRequest struct {
	PaymentID uint   `json:"payment_id" binding:"required"`
	Authority string `json:"authority" binding:"required"` // From Zarinpal
}

type VerifyResponse struct {
	PaymentID       uint   `json:"payment_id"`
	Status          string `json:"status"`
	ReferenceNumber string `json:"reference_number"`
	Message         string `json:"message"`
}

// Verify verifies payment with Zarinpal
func (s *PaymentService) Verify(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error) {
	s.logger.Info("Verifying payment",
		zap.Uint("payment_id", req.PaymentID),
		zap.String("authority", req.Authority),
	)

	// 1. Get payment record
	payment, err := s.paymentRepo.GetByID(ctx, req.PaymentID)
	if err != nil {
		s.logger.Warn("Payment not found", zap.Uint("payment_id", req.PaymentID))
		return nil, errors.New("payment not found")
	}

	if payment.Status != domain.PaymentVerifying {
		s.logger.Warn("Invalid payment status for verification",
			zap.Uint("payment_id", req.PaymentID),
			zap.String("status", string(payment.Status)),
		)
		return &VerifyResponse{
			PaymentID: payment.ID,
			Status:    string(payment.Status),
			Message:   "invalid payment status",
		}, nil
	}

	// 2. Verify with Zarinpal
	zarinpalResp, err := s.zarinpalClient.VerifyPayment(payment.Amount, req.Authority)
	if err != nil {
		s.logger.Error("Zarinpal verification failed", zap.Error(err))
		payment.Status = domain.PaymentFailed
		payment.ErrorMessage = err.Error()
		s.paymentRepo.UpdateStatus(ctx, payment.ID, domain.PaymentFailed)
		return &VerifyResponse{
			PaymentID: payment.ID,
			Status:    string(domain.PaymentFailed),
			Message:   "verification failed",
		}, nil
	}

	// 3. Update payment status
	payment.Status = domain.PaymentCompleted
	now := time.Now()
	payment.VerifiedAt = &now
	payment.ReferenceNumber = fmt.Sprintf("%d", zarinpalResp.Data.RefID)

	if err := s.paymentRepo.UpdateVerification(ctx, payment.ID, domain.PaymentCompleted, fmt.Sprintf("%d", zarinpalResp.Data.RefID)); err != nil {
		s.logger.Error("Failed to update payment verification", zap.Error(err))
		return nil, err
	}

	// 4. Update order status
	order, err := s.orderRepo.GetByID(ctx, payment.OrderID)
	if err != nil {
		s.logger.Error("Failed to get order for payment", zap.Uint("order_id", payment.OrderID), zap.Error(err))
		return nil, err
	}

	if err := s.orderRepo.UpdateStatus(ctx, order.ID, string(domain.OrderPaid)); err != nil {
		s.logger.Error("Failed to update order status", zap.Uint("order_id", order.ID), zap.Error(err))
		return nil, err
	}

	// 5. Add funds to user wallet (if needed for future purchases)
	// In this case, we're just marking order as paid

	s.logger.Info("Payment verification successful",
		zap.Uint("payment_id", payment.ID),
		zap.String("reference_number", payment.ReferenceNumber),
	)

	return &VerifyResponse{
		PaymentID:       payment.ID,
		Status:          string(domain.PaymentCompleted),
		ReferenceNumber: payment.ReferenceNumber,
		Message:         "payment verified successfully",
	}, nil
}
