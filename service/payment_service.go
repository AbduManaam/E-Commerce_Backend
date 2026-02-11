package service

import (
	"backend/internal/domain"
	"backend/repository"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

var (
	ErrPaymentAlreadyExists = errors.New("payment already exists for this order")
	ErrOrderNotPayable      = errors.New("order is not payable")
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

func (s *PaymentService) CreatePaymentIntent(
	orderID uint,
	method domain.PaymentMethod,
) (*domain.Payment, string, error) {

	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, "", err
	}

	if order.Status != domain.OrderStatusPending {
		return nil, "", ErrOrderNotPayable
	}

	// Check if payment already exists
	if existingPayment, err := s.paymentRepo.GetByOrderID(orderID); err == nil {
		 log.Printf("Payment already exists: %+v", existingPayment)
		return nil, "", ErrPaymentAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	payment := &domain.Payment{
		OrderID:       orderID,
		PaymentMethod: method,
		Amount:        order.FinalTotal,
		Currency:      "INR",
		Status:        domain.PaymentStatusPending,
		GatewayData:   json.RawMessage(`{}`),
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, "", err
	}

	var clientSecret string
	switch method {
	case domain.PaymentMethodRazorpay:
		payment.GatewayID = fmt.Sprintf("razorpay_order_%d", payment.ID)
		clientSecret = payment.GatewayID
	case domain.PaymentMethodStripe:
		payment.GatewayID = fmt.Sprintf("stripe_payment_%d", payment.ID)
		clientSecret = payment.GatewayID
	case domain.PaymentMethodCOD:
		// For COD, mark as paid immediately
		now := time.Now().UTC()
		payment.Status = domain.PaymentStatusPaid
		payment.PaidAt = &now
	}

	if err := s.paymentRepo.Update(payment); err != nil {
		return nil, "", err
	}

	return payment, clientSecret, nil
}



func (s *PaymentService) ConfirmPayment(paymentID string, status string) (*domain.Payment, error) {
    // Fetch payment by GatewayID
    payment, err := s.paymentRepo.GetByGatewayID(paymentID)
    if err != nil {
        s.logger.Printf("ConfirmPayment: payment not found for gatewayID=%s", paymentID)
        return nil, err
    }

    // Ignore if already paid
    if payment.Status == domain.PaymentStatusPaid {
        s.logger.Printf("ConfirmPayment: payment already marked as paid, paymentID=%d", payment.ID)
        return payment, nil
    }

    now := time.Now().UTC()
    switch status {
    case "success":
        payment.Status = domain.PaymentStatusPaid
        payment.PaidAt = &now
        s.logger.Printf("ConfirmPayment: payment success, paymentID=%d, amount=%f", payment.ID, payment.Amount)
    case "failed":
        payment.Status = domain.PaymentStatusFailed
        s.logger.Printf("ConfirmPayment: payment failed, paymentID=%d", payment.ID)
    default:
        return nil, errors.New("invalid payment status")
    }

    if err := s.paymentRepo.Update(payment); err != nil {
        s.logger.Printf("ConfirmPayment: failed to update payment, paymentID=%d, error=%s", payment.ID, err)
        return nil, err
    }

    return payment, nil
}
