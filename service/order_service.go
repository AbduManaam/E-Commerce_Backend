package service

import (
	"backend/internal/domain"
	"backend/repository"
	"errors"
	"log"
	"time"
)

type OrderService struct {
	orderRepo    repository.OrderRepository
	productRead  repository.ProductReader
	productWrite repository.ProductWriter
	cartRepo     repository.CartRepository
	logger       *log.Logger
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	productRead repository.ProductReader,
	productWrite repository.ProductWriter,
	cartRepo repository.CartRepository,
	logger *log.Logger,
) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		productRead:  productRead,
		productWrite: productWrite,
		cartRepo:     cartRepo,
		logger:       logger,
	}
}

func (s *OrderService) CreateOrder(
	userID uint,
	addressID uint,
	paymentMethod domain.PaymentMethod,
) (*domain.Order, error) {

	tx := s.orderRepo.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Get cart
	cart, err := s.cartRepo.GetForUpdate(tx, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(cart.Items) == 0 {
		tx.Rollback()
		return nil, ErrCartEmpty
	}

	var (
		orderItems    []domain.OrderItem
		total         float64
		totalDiscount float64
		now           = time.Now()
	)

	// 2. Process each cart item
	for _, ci := range cart.Items {

		product, err := s.productRepo.GetByIDForUpdate(tx, ci.ProductID)
		if err != nil {
			tx.Rollback()
			return nil, ErrProductNotFound
		}

		if !product.IsActive || product.Stock <= 0 {
			tx.Rollback()
			return nil, ErrProductUnavailable
		}

		if product.Stock < int(ci.Quantity) {
			tx.Rollback()
			return nil, ErrInsufficientStock
		}

		unitPrice := product.Price
		finalPrice := product.CalculatePrice(now)
		discountPerUnit := unitPrice - finalPrice

		subtotal := finalPrice * float64(ci.Quantity)

		orderItems = append(orderItems, domain.OrderItem{
			ProductID:       product.ID,
			Quantity:        ci.Quantity,
			Price:           unitPrice,
			DiscountPercent: discountPerUnit,
			FinalPrice:      finalPrice,
			Subtotal:        subtotal,
			Status:          domain.OrderItemStatusPending,
		})

		total += unitPrice * float64(ci.Quantity)
		totalDiscount += discountPerUnit * float64(ci.Quantity)

		// Reduce stock
		product.Stock -= int(ci.Quantity)
		if err := s.productRepo.UpdateTx(tx, product); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 3. Create order
	order := &domain.Order{
		UserID:            userID,
		Items:             orderItems,
		Total:             total,
		Discount:          totalDiscount,
		FinalTotal:        total - totalDiscount,
		Status:            domain.OrderStatusPending,
		ShippingAddressID: &addressID,
		PaymentMethod:     paymentMethod,
		PaymentStatus:     domain.PaymentStatusPending,
	}

	if err := s.orderRepo.CreateTx(tx, order); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 4. Clear cart
	if err := s.cartRepo.ClearTx(tx, userID); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 5. Commit
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) GetUserOrders(userID uint) ([]domain.Order, error) {
	if userID == 0 {
		s.logger.Printf("GetUserOrders failed: invalid userID=0")
		return nil, errors.New("invalid user id")
	}

	orders, err := s.orderRepo.GetOrdersByUserID(userID)
	if err != nil {
		s.logger.Printf(
			"GetUserOrders failed: repo error userID=%d err=%v",
			userID, err,
		)
		return nil, err
	}

	if len(orders) == 0 {
		return []domain.Order{}, nil
	}

	return orders, nil
}

func (s *OrderService) GetOrder(userID uint, orderID uint) (*domain.Order, error) {
	if userID == 0 || orderID == 0 {
		s.logger.Printf(
			"GetOrder failed: invalid input userID=%d orderID=%d",
			userID, orderID,
		)
		return nil, ErrInvalidInput
	}

	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		s.logger.Printf(
			"GetOrder failed: order not found orderID=%d err=%v",
			orderID, err,
		)
		return nil, err
	}

	if order.UserID != userID {
		s.logger.Printf(
			"GetOrder forbidden: userID=%d orderID=%d ownerID=%d",
			userID, orderID, order.UserID,
		)
		return nil, ErrForbidden
	}

	return order, nil
}

func (s *OrderService) CancelOrder(userID uint, orderID uint) error {
	if userID == 0 || orderID == 0 {
		s.logger.Printf(
			"CancelOrder failed: invalid input userID=%d orderID=%d",
			userID, orderID,
		)
		return ErrInvalidInput
	}

	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		s.logger.Printf(
			"CancelOrder failed: order not found orderID=%d err=%v",
			orderID, err,
		)
		return err
	}

	if order.UserID != userID {
		s.logger.Printf(
			"CancelOrder forbidden: userID=%d orderID=%d ownerID=%d",
			userID, orderID, order.UserID,
		)
		return ErrForbidden
	}

	if order.Status != domain.OrderStatusPending {
		s.logger.Printf(
			"CancelOrder failed: invalid status orderID=%d status=%s",
			orderID, order.Status,
		)
		return ErrOrderNotCancelable
	}

	return s.orderRepo.UpdateStatus(orderID, domain.OrderStatusCanceled)
}

func (s *OrderService) ListAllOrders() ([]domain.Order, error) {
	orders, err := s.orderRepo.ListAll()
	if err != nil {
		s.logger.Printf("ListAllOrders failed: err=%v", err)
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) UpdateOrderStatus(orderID uint, status domain.OrderStatus) error {
	if orderID == 0 {
		s.logger.Printf("UpdateOrderStatus failed: invalid orderID=0")
		return ErrInvalidInput
	}

	if !domain.IsValidOrderStatus(status) {
		s.logger.Printf(
			"UpdateOrderStatus failed: invalid status orderID=%d status=%s",
			orderID, status,
		)
		return ErrInvalidOrderStatus
	}

	return s.orderRepo.UpdateStatus(orderID, status)
}
