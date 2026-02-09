package service

import (
	"backend/internal/domain"
	"backend/repository"
	"errors"
	"log"
	"time"
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

	if _, err := s.paymentRepo.GetByOrderID(orderID); err == nil {
		return nil, "", ErrPaymentAlreadyExists
	}

	payment := &domain.Payment{
		OrderID:       orderID,
		PaymentMethod: method,
		Amount:        order.FinalTotal,
		Currency:      "INR",
		Status:        domain.PaymentStatusPending,
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, "", err
	}

	switch method {
	case domain.PaymentMethodRazorpay:
		payment.GatewayID = "razorpay_order_id"
	case domain.PaymentMethodStripe:
		payment.GatewayID = "stripe_payment_intent_id"
	case domain.PaymentMethodCOD:
		now := time.Now().UTC()
		payment.Status = domain.PaymentStatusPaid
		payment.PaidAt = &now
	}

	_ = s.paymentRepo.Update(payment)

	return payment, payment.GatewayID, nil
}
