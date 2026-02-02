package service

import (
	"backend/handler/dto"
	"backend/internal/domain"
	"backend/repository"
	"errors"
)

type OrderService struct {
	orderRepo repository.OrderRepository
	productRepo repository.ProductReader
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	productRepo repository.ProductReader,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}


func (s *OrderService) CreateOrder(
    userID uint,
    req dto.CreateOrderRequest,
) (*domain.Order, error) {

    var orderItems []domain.OrderItem
    var total float64

    for _, item := range req.Items { // item is OrderItemRequest
        if item.Quantity <= 0 {
            return nil, ErrInvalidInput
        }

        product, err := s.productRepo.GetByID(item.ProductID)
        if err != nil {
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
        Status: "pending",
        Total:  total,
        Items:  orderItems,
    }

    if err := s.orderRepo.Create(order); err != nil {
        return nil, err
    }

    return order, nil
}


func (s *OrderService) GetUserOrders(userID uint) ([]domain.Order, error) {
    if userID == 0 {
        return nil, errors.New("invalid user id")
    }

    orders, err := s.orderRepo.GetOrdersByUserID(userID)
    if err != nil {
        return nil, err
    }

    if len(orders) == 0 {
        return []domain.Order{}, nil
    }

    return orders, nil
}

func (s *OrderService) GetOrder(userID uint, orderID uint) (*domain.Order, error) {
	if userID == 0 || orderID == 0 {
		return nil, ErrInvalidInput
	}

	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, ErrForbidden
	}

	return order, nil
}

func (s *OrderService) CancelOrder(userID uint, orderID uint) error {
	if userID == 0 || orderID == 0 {
		return ErrInvalidInput
	}

	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return err
	}

	if order.UserID != userID {
		return ErrForbidden
	}

	if order.Status != domain.OrderStatusPending {
		return ErrOrderNotCancelable
	}

	return s.orderRepo.UpdateStatus(orderID, domain.OrderStatusCanceled)
}


func (s *OrderService) ListAllOrders() ([]domain.Order, error) {
	return s.orderRepo.ListAll()
}

func (s *OrderService) UpdateOrderStatus(orderID uint, status domain.OrderStatus) error {
	if orderID == 0 {
		return ErrInvalidInput
	}

	if !domain.IsValidOrderStatus(status) {
		return ErrInvalidOrderStatus
	}

	return s.orderRepo.UpdateStatus(orderID, status)
}

