

package service

import (
	"backend/internal/domain"
	"backend/repository"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type PaymentService struct {
	paymentRepo repository.PaymentRepository
	orderRepo   repository.OrderRepository
	logger      *log.Logger
}

func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	orderRepo repository.OrderRepository,
	logger *log.Logger,
) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		logger:      logger,
	}
}


// CreatePaymentIntent remains the same
func (s *PaymentService) CreatePaymentIntent(
	orderID uint,
	method domain.PaymentMethod,
) (*domain.Payment, string, error) {
	// Start a transaction
	tx := s.paymentRepo.GetDB().Begin()

	// Fetch the order
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", fmt.Errorf("order %d not found", orderID)
		}
		return nil, "", err
	}

	// Check if order is payable
	if order.Status != domain.OrderStatusPending {
		tx.Rollback()
		return nil, "", errors.New("order is not payable")
	}

	// Check if payment already exists
	if existingPayment, err := s.paymentRepo.GetByOrderID(orderID); err == nil {
		tx.Rollback()
		s.logger.Printf("Payment already exists: %+v", existingPayment)
		return nil, "", errors.New("payment already exists for this order")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, "", err
	}

	// Create new payment
	payment := &domain.Payment{
		OrderID:       orderID,
		PaymentMethod: method,
		Amount:        order.FinalTotal,
		Currency:      "INR",
		Status:        domain.PaymentStatusPending,
	}

	if err := s.paymentRepo.CreateTx(tx, payment); err != nil {
		tx.Rollback()
		return nil, "", err
	}

	var clientSecret string
	now := time.Now().UTC()

	switch method {
	case domain.PaymentMethodRazorpay:
		payment.GatewayID = fmt.Sprintf("razorpay_order_%d", payment.ID)
		clientSecret = payment.GatewayID

	case domain.PaymentMethodStripe:
		payment.GatewayID = fmt.Sprintf("stripe_payment_%d", payment.ID)
		clientSecret = payment.GatewayID

	case domain.PaymentMethodCOD:
		// COD: mark payment and order as paid immediately
		payment.Status = domain.PaymentStatusPaid
		payment.PaidAt = &now

		order.PaymentStatus = domain.PaymentStatusPaid
		order.PaidAt = &now
		order.PaymentMethod = method
	}

	// Update payment & order
	if err := s.paymentRepo.UpdateTx(tx, payment); err != nil {
		tx.Rollback()
		return nil, "", err
	}

	if method == domain.PaymentMethodCOD {
		if err := s.orderRepo.UpdateTx(tx, order); err != nil {
			tx.Rollback()
			return nil, "", err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, "", err
	}

	s.logger.Printf("PaymentIntent created successfully: paymentID=%d, orderID=%d, method=%s", payment.ID, order.ID, method)
	return payment, clientSecret, nil
}


// ------------------- ConfirmPayment Updated -------------------

func (s *PaymentService) ConfirmPayment(paymentID string, status string) (*domain.Payment, error) {
	// Start a DB transaction
	tx := s.paymentRepo.GetDB().Begin() // Assume GetDB() returns *gorm.DB

	payment, err := s.paymentRepo.GetByGatewayID(paymentID)
	if err != nil {
		tx.Rollback()
		s.logger.Printf("ConfirmPayment: payment not found for gatewayID=%s", paymentID)
		return nil, err
	}

	if payment.Status == domain.PaymentStatusPaid {
		tx.Rollback()
		s.logger.Printf("ConfirmPayment: payment already marked as paid, paymentID=%d", payment.ID)
		return payment, nil
	}

	now := time.Now().UTC()
	switch status {
	case "success":
		payment.Status = domain.PaymentStatusPaid
		payment.PaidAt = &now
	case "failed":
		payment.Status = domain.PaymentStatusFailed
	default:
		tx.Rollback()
		return nil, errors.New("invalid payment status")
	}

	// Update Payment
	if err := s.paymentRepo.UpdateTx(tx, payment); err != nil {
		tx.Rollback()
		s.logger.Printf("ConfirmPayment: failed to update payment, paymentID=%d, error=%s", payment.ID, err)
		return nil, err
	}

	// Update corresponding Order
	order, err := s.orderRepo.GetByID(payment.OrderID)
	if err != nil {
		tx.Rollback()
		s.logger.Printf("ConfirmPayment: order not found for orderID=%d", payment.OrderID)
		return nil, err
	}

	if payment.Status == domain.PaymentStatusPaid {
		order.PaymentStatus = domain.PaymentStatusPaid
		order.PaymentMethod = payment.PaymentMethod
		order.PaidAt = payment.PaidAt
	}

	if err := s.orderRepo.UpdateTx(tx, order); err != nil {
		tx.Rollback()
		s.logger.Printf("ConfirmPayment: failed to update order, orderID=%d, error=%s", order.ID, err)
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		s.logger.Printf("ConfirmPayment: failed to commit transaction, paymentID=%d, error=%s", payment.ID, err)
		return nil, err
	}

	s.logger.Printf("ConfirmPayment: payment & order updated successfully, paymentID=%d, orderID=%d", payment.ID, order.ID)
	return payment, nil
}
