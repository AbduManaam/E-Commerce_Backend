package service

import (
	"backend/handler/dto"
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
	cartRepo     repository.CartRepositoryInterface
	addressRepo  repository.AddressRepository 
	logger       *log.Logger
	
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	productRead repository.ProductReader,
	productWrite repository.ProductWriter,
	cartRepo repository.CartRepositoryInterface,
	addressRepo repository.AddressRepository,
	logger *log.Logger,
) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		productRead:  productRead,
		productWrite: productWrite,
		cartRepo:     cartRepo,
		addressRepo:  addressRepo,
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

	// ✅ Step 1: Fetch user address
	address, err := s.addressRepo.GetByID(userID, addressID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// ✅ Step 2: Convert Address → OrderAddress (SNAPSHOT)
	orderAddress := &domain.OrderAddress{
		FullName: address.FullName,
		Phone:    address.Phone,
		Address:  address.Address,
		City:     address.City,
		State:    address.State,
		Country:  address.Country,
		ZipCode:  address.ZipCode,
		Landmark: address.Landmark,
	}

	if err := tx.Create(orderAddress).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 3️⃣ Get cart (locked)
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

	// 4️⃣ Process cart items
	for _, ci := range cart.Items {
		product, err := s.productRead.GetByIDForUpdate(tx, ci.ProductID)
		if err != nil {
			tx.Rollback()
			return nil, ErrProductNotFound
		}

		if !product.IsActive || product.Stock < int(ci.Quantity) {
			tx.Rollback()
			return nil, ErrProductUnavailable
		}

		unitPrice := product.Price
		finalPrice := product.CalculatePrice(now)
		discount := unitPrice - finalPrice
		subtotal := finalPrice * float64(ci.Quantity)

		orderItems = append(orderItems, domain.OrderItem{
			ProductID:       product.ID,
			Quantity:        ci.Quantity,
			Price:           unitPrice,
			DiscountAmount:  discount,
			FinalPrice:      finalPrice,
			Subtotal:        subtotal,
			Status:          domain.OrderItemStatusPending,
		})

		total += unitPrice * float64(ci.Quantity)
		totalDiscount += discount * float64(ci.Quantity)

		product.Stock -= int(ci.Quantity)
		if err := s.productWrite.UpdateTx(tx, product); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 5️⃣ Create order (VALID FK)
	order := &domain.Order{
		UserID:            userID,
		Items:             orderItems,
		Total:             total,
		Discount:          totalDiscount,
		FinalTotal:        total - totalDiscount,
		Status:            domain.OrderStatusPending,
		ShippingAddressID: &orderAddress.ID,
		PaymentMethod:     paymentMethod,
		PaymentStatus:     domain.PaymentStatusPending,
	}

	if err := s.orderRepo.CreateTx(tx, order); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 6️⃣ Clear cart
	if err := s.cartRepo.ClearTx(tx, userID); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	// After tx.Commit().Error
fullOrder, err := s.orderRepo.GetByIDWithAssociations(order.ID)
if err != nil {
    return nil, err
}

return fullOrder, nil


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

	return s.orderRepo.UpdateStatus(orderID, domain.OrderStatusCancelled)
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



func (s *OrderService) CancelOrderItem(userID, orderID, itemID uint, reason string) error {
    tx := s.orderRepo.Begin() // start transaction
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // 1️⃣ Get order with items (lock for update)
    order, err := s.orderRepo.GetByID(orderID)
    if err != nil {
        tx.Rollback()
        return ErrNotFound
    }

    // 2️⃣ Verify ownership
    if order.UserID != userID {
        tx.Rollback()
        return ErrForbidden
    }

    // 3️⃣ Prevent cancelling if shipped/delivered
    if order.Status == domain.OrderStatusShipped || order.Status == domain.OrderStatusDelivered {
        tx.Rollback()
        return &ServiceError{
            Code: "ORDER_ALREADY_PROCESSED",
            Msg:  "Cannot cancel items for shipped/delivered orders",
        }
    }

    // 4️⃣ Find item
    var itemToCancel *domain.OrderItem
    for i := range order.Items {
        if order.Items[i].ID == itemID {
            itemToCancel = &order.Items[i]
            break
        }
    }
    if itemToCancel == nil {
        tx.Rollback()
        return ErrNotFound
    }

    // 5️⃣ Check if item can be cancelled
    if !itemToCancel.CanBeCancelled() {
        tx.Rollback()
        return &ServiceError{
            Code: "ITEM_NOT_CANCELLABLE",
            Msg:  "This item cannot be cancelled",
        }
    }

    // 6️⃣ Update item
    now := time.Now().UTC()
    itemToCancel.Status = domain.OrderItemStatusCancelled
    itemToCancel.CancellationReason = &reason
    itemToCancel.CancelledAt = &now

    if err := s.orderRepo.UpdateOrderItemTx(tx, itemToCancel); err != nil {
        tx.Rollback()
        return err
    }

    // 7️⃣ Restore stock
    product, err := s.productRead.GetByIDForUpdate(tx, itemToCancel.ProductID)
    if err == nil && product != nil {
        product.Stock += int(itemToCancel.Quantity)
        if err := s.productWrite.UpdateTx(tx, product); err != nil {
            tx.Rollback()
            return err
        }
    }

    // 8️⃣ Recalculate order totals & update status
    allCancelled := true
    newTotal := 0.0
    for _, item := range order.Items {
        if item.Status != domain.OrderItemStatusCancelled {
            allCancelled = false
            newTotal += item.FinalPrice * float64(item.Quantity)
        }
    }
    order.FinalTotal = newTotal
    if allCancelled {
        order.Status = domain.OrderStatusCancelled
    }

    // 9️⃣ Save order
    if err := s.orderRepo.UpdateTx(tx, order); err != nil {
        tx.Rollback()
        return err
    }

    // 10️⃣ Commit transaction
    if err := tx.Commit().Error; err != nil {
        return err
    }

    return nil
}

func (s *OrderService) GetUserOrdersPaginated(
	userID uint,
	page int,
	limit int,
) ([]domain.Order, int64, error) {

	if userID == 0 {
		return nil, 0, ErrInvalidInput
	}

	offset := (page - 1) * limit

	orders, err := s.orderRepo.GetOrdersByUserIDPaginated(userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.orderRepo.CountOrdersByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (s *OrderService) GetOrderDetail(orderID uint) (*domain.Order, error) {
	if orderID == 0 {
		return nil, ErrInvalidInput
	}

	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, err
	}

	return order, nil
}




func (s *OrderService) CreateOrderFromCart(
	userID uint,
	addressID uint,
	paymentMethod domain.PaymentMethod,
) (*domain.Order, error) {
	// Use your CreateOrder service, which now reloads associations
	return s.CreateOrder(userID, addressID, paymentMethod)
}

func (s *OrderService) CreateDirectOrder(
	userID uint,
	addressID uint,
	paymentMethod domain.PaymentMethod,
	items []dto.OrderItemRequest,
) (*domain.Order, error) {

	tx := s.orderRepo.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Address → OrderAddress
	address, err := s.addressRepo.GetByID(userID, addressID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	orderAddress := &domain.OrderAddress{
		FullName: address.FullName,
		Phone:    address.Phone,
		Address:  address.Address,
		City:     address.City,
		State:    address.State,
		Country:  address.Country,
		ZipCode:  address.ZipCode,
		Landmark: address.Landmark,
	}

	if err := tx.Create(orderAddress).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var (
		orderItems    []domain.OrderItem
		total         float64
		totalDiscount float64
		now           = time.Now()
	)

	for _, req := range items {
		product, err := s.productRead.GetByIDForUpdate(tx, req.ProductID)
		if err != nil {
			tx.Rollback()
			return nil, ErrProductNotFound
		}

		unitPrice := product.Price
		finalPrice := product.CalculatePrice(now)
		discount := unitPrice - finalPrice
		subtotal := finalPrice * float64(req.Quantity)

		orderItems = append(orderItems, domain.OrderItem{
			ProductID:       product.ID,
			Quantity:        req.Quantity,
			Price:           unitPrice,
			DiscountAmount:  discount,
			FinalPrice:      finalPrice,
			Subtotal:        subtotal,
			Status:          domain.OrderItemStatusPending,
		})

		total += unitPrice * float64(req.Quantity)
		totalDiscount += discount * float64(req.Quantity)

		product.Stock -= int(req.Quantity)
		if err := s.productWrite.UpdateTx(tx, product); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	order := &domain.Order{
		UserID:            userID,
		Items:             orderItems,
		Total:             total,
		Discount:          totalDiscount,
		FinalTotal:        total - totalDiscount,
		Status:            domain.OrderStatusPending,
		ShippingAddressID: &orderAddress.ID,
		PaymentMethod:     paymentMethod,
		PaymentStatus:     domain.PaymentStatusPending,
	}

	if err := s.orderRepo.CreateTx(tx, order); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// -------------------------
	// Reload with associations
	// -------------------------
	fullOrder, err := s.orderRepo.GetByIDWithAssociations(order.ID)
	if err != nil {
		return nil, err
	}

	return fullOrder, nil
}
