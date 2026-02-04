package service

import (
	"backend/handler/dto"
	"backend/internal/domain"
	"backend/repository"
	"errors"
	"log"
)

type OrderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductReader
	logger      *log.Logger
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	productRepo repository.ProductReader,
	logger *log.Logger,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		logger:      logger,
	}
}

func (s *OrderService) CreateOrder(
	userID uint,
	req dto.CreateOrderRequest,
) (*domain.Order, error) {

	var orderItems []domain.OrderItem
	var total float64

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			s.logger.Printf(
				"CreateOrder failed: invalid quantity userID=%d productID=%d qty=%d",
				userID, item.ProductID, item.Quantity,
			)
			return nil, ErrInvalidInput
		}

		product, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			s.logger.Printf(
				"CreateOrder failed: product not found userID=%d productID=%d err=%v",
				userID, item.ProductID, err,
			)
			return nil, ErrProductNotFound
		}

		lineTotal := product.Price * float64(item.Quantity)
		total += lineTotal

		orderItems = append(orderItems, domain.OrderItem{
			ProductID: product.ID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		})
	}

	order := &domain.Order{
		UserID: userID,
		Status: domain.OrderStatusPending,
		Total:  total,
		Items:  orderItems,
	}

	if err := s.orderRepo.Create(order); err != nil {
		s.logger.Printf(
			"CreateOrder failed: db create error userID=%d total=%.2f err=%v",
			userID, total, err,
		)
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
