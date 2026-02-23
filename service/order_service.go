package service

import (
	"backend/handler/dto"
	"backend/internal/domain"
	"backend/repository"
	"errors"
	"log/slog"
	"time"
)

type OrderService struct {
	orderRepo    repository.OrderRepository
	productRead  repository.ProductReader
	productWrite repository.ProductWriter
	cartRepo     repository.CartRepositoryInterface
	addressRepo  repository.AddressRepository
	paymentSvc   *PaymentService
	logger       *slog.Logger
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	productRead repository.ProductReader,
	productWrite repository.ProductWriter,
	cartRepo repository.CartRepositoryInterface,
	addressRepo repository.AddressRepository,
	paymentSvc *PaymentService,
	logger *slog.Logger,
) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		productRead:  productRead,
		productWrite: productWrite,
		cartRepo:     cartRepo,
		addressRepo:  addressRepo,
		paymentSvc:   paymentSvc,
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

	// Fetch user address
	address, err := s.addressRepo.GetByID(userID, addressID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Convert Address → OrderAddress (SNAPSHOT)
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

	// Get cart (locked)
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

	// Process cart items
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

		// ✅ FIXED: Get price from product_prices table
		var unitPrice float64
		var finalPrice float64

		if len(product.Prices) > 0 {
			// Try to get "H" (Half) price first
			price, err := product.GetPriceByType("H")
			if err != nil {
				// If "H" not found, use first available price
				if len(product.Prices) > 0 {
					unitPrice = product.Prices[0].Price
				} else {
					tx.Rollback()
					return nil, ErrPriceNotFound
				}
			} else {
				unitPrice = price
			}

			// Calculate final price with any active discounts
			finalPrice = product.CalculatePrice("H", now)
			if finalPrice == 0 {
				finalPrice = unitPrice
			}
		} else {
			// No prices found for this product
			tx.Rollback()
			return nil, ErrPriceNotFound
		}

		discount := unitPrice - finalPrice
		subtotal := finalPrice * float64(ci.Quantity)

		orderItems = append(orderItems, domain.OrderItem{
			ProductID:      product.ID,
			Quantity:       ci.Quantity,
			Price:          unitPrice,
			DiscountAmount: discount,
			FinalPrice:     finalPrice,
			Subtotal:       subtotal,
			Status:         domain.OrderItemStatusPending,
		})

		total += unitPrice * float64(ci.Quantity)
		totalDiscount += discount * float64(ci.Quantity)

		product.Stock -= int(ci.Quantity)
		if err := s.productWrite.UpdateTx(tx, product); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Create order
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

	// Clear cart
	if err := s.cartRepo.ClearTx(tx, userID); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Reload with associations
	fullOrder, err := s.orderRepo.GetByIDWithAssociations(order.ID)
	if err != nil {
		return nil, err
	}

	return fullOrder, nil
}
func (s *OrderService) GetUserOrders(userID uint) ([]dto.OrderResponse, error) {
	if userID == 0 {
		s.logger.Warn("GetUserOrders failed: invalid userID", "user_id", 0)
		return nil, errors.New("invalid user id")
	}

	orders, err := s.orderRepo.GetOrdersByUserID(userID)
	if err != nil {
		s.logger.Error("GetUserOrders failed: repo error", "user_id", userID, "error", err)
		return nil, err
	}

	if len(orders) == 0 {
		return []dto.OrderResponse{}, nil
	}

	var response []dto.OrderResponse

	for _, order := range orders {

		var items []dto.OrderItemResponse

		for _, item := range order.Items {
			var images []dto.ProductImageDTO
			for _, img := range item.Product.Images {
				images = append(images, dto.ProductImageDTO{URL: img.URL})
			}

			items = append(items, dto.OrderItemResponse{
				ID:             item.ID,
				Quantity:       item.Quantity,
				Price:          item.Price,
				DiscountAmount: item.DiscountAmount,
				FinalPrice:     item.FinalPrice,
				Subtotal:       item.Subtotal,
				Status:         string(item.Status),
				Product: dto.ProductResponse{
					ID:     item.Product.ID,
					Name:   item.Product.Name,
					Price:  item.Product.FinalPrice,
					Images: images,
				},
			})
		}

		// Map shipping address
		var shippingAddress *dto.OrderAddressDTO
		if order.ShippingAddress != nil {
			shippingAddress = &dto.OrderAddressDTO{
				FullName: order.ShippingAddress.FullName,
				Phone:    order.ShippingAddress.Phone,
				Address:  order.ShippingAddress.Address,
				City:     order.ShippingAddress.City,
				State:    order.ShippingAddress.State,
				Country:  order.ShippingAddress.Country,
				ZipCode:  order.ShippingAddress.ZipCode,
				Landmark: order.ShippingAddress.Landmark,
			}
		}

		response = append(response, dto.OrderResponse{
			ID:              order.ID,
			Status:          string(order.Status),
			Total:           order.Total,
			Discount:        order.Discount,
			FinalTotal:      order.FinalTotal,
			PaymentMethod:   string(order.PaymentMethod),
			PaymentStatus:   string(order.PaymentStatus),
			CreatedAt:       order.CreatedAt,
			ShippingAddress: shippingAddress,
			Items:           items,
		})
	}

	return response, nil
}

func (s *OrderService) GetOrder(userID uint, orderID uint) (*domain.Order, error) {
	if userID == 0 || orderID == 0 {
		s.logger.Warn("GetOrder failed: invalid input", "user_id", userID, "order_id", orderID)
		return nil, ErrInvalidInput
	}

	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		s.logger.Error("GetOrder failed: order not found", "order_id", orderID, "error", err)
		return nil, err
	}

	if order.UserID != userID {
		s.logger.Warn("GetOrder forbidden", "user_id", userID, "order_id", orderID, "owner_id", order.UserID)
		return nil, ErrForbidden
	}

	return order, nil
}

func (s *OrderService) CancelOrder(userID uint, orderID uint) error {
	if userID == 0 || orderID == 0 {
		s.logger.Warn("CancelOrder failed: invalid input", "user_id", userID, "order_id", orderID)
		return ErrInvalidInput
	}

	// 👇 IMPORTANT: preload items
	order, err := s.orderRepo.GetByIDWithItems(orderID)
	if err != nil {
		s.logger.Error("CancelOrder failed: order not found", "order_id", orderID, "error", err)
		return err
	}

	if order.UserID != userID {
		s.logger.Warn("CancelOrder forbidden", "user_id", userID, "order_id", orderID, "owner_id", order.UserID)
		return ErrForbidden
	}

	if order.Status != domain.OrderStatusPending {
		s.logger.Warn("CancelOrder failed: invalid status", "order_id", orderID, "status", order.Status)
		return ErrOrderNotCancelable
	}

	// ✅ update order status
	order.Status = domain.OrderStatusCancelled

	// ✅ update ALL item statuses
	for i := range order.Items {
		order.Items[i].Status = domain.OrderItemStatusCancelled
	}

	// ✅ save everything
	if err := s.orderRepo.SaveOrderWithItems(order); err != nil {
		s.logger.Error("CancelOrder failed: save error", "order_id", orderID, "error", err)
		return err
	}

	// ✅ trigger refund for online payments (Razorpay/Stripe) - RefundPayment checks Payment.Status internally
	if order.PaymentMethod != domain.PaymentMethodCOD && s.paymentSvc != nil {
		if err := s.paymentSvc.RefundPayment(orderID); err != nil {
			s.logger.Error("CancelOrder: refund failed",
				"order_id", orderID, "error", err,
			)
		}
	}

	return nil
}

func (s *OrderService) ListAllOrders() ([]domain.Order, error) {
	orders, err := s.orderRepo.ListAll()
	if err != nil {
		s.logger.Error("ListAllOrders failed", "error", err)
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) UpdateOrderStatus(orderID uint, status domain.OrderStatus) error {
	if orderID == 0 {
		s.logger.Warn("UpdateOrderStatus failed: invalid orderID", "order_id", 0)
		return ErrInvalidInput
	}

	if !domain.IsValidOrderStatus(status) {
		s.logger.Warn("UpdateOrderStatus failed: invalid status", "order_id", orderID, "status", status)
		return ErrInvalidOrderStatus
	}

	// When admin cancels a Razorpay paid order, process the actual refund
	if status == domain.OrderStatusCancelled && s.paymentSvc != nil {
		order, err := s.orderRepo.GetByID(orderID)
		if err == nil && order.PaymentMethod == domain.PaymentMethodRazorpay && order.PaymentStatus == domain.PaymentStatusPaid {
			if err := s.paymentSvc.RefundPayment(orderID); err == nil {
				s.logger.Info("UpdateOrderStatus: auto-refunded Razorpay order", "order_id", orderID)
				return nil // RefundPayment already updated order status to "refunded"
			}
			// Refund failed (e.g. already refunded); fall through to UpdateStatus
		}
	}

	return s.orderRepo.UpdateStatus(orderID, status)
}

func (s *OrderService) CancelOrderItem(
	userID, orderID, itemID uint,
	reason string,
) error {

	tx := s.orderRepo.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	order, err := s.orderRepo.GetByIDForUpdate(tx, orderID)
	if err != nil {
		tx.Rollback()
		return ErrNotFound
	}

	s.logger.Debug("CancelOrderItem: ownership check", "order_user_id", order.UserID, "request_user_id", userID)

	if order.UserID != userID {
		tx.Rollback()
		return ErrForbidden
	}

	if order.Status == domain.OrderStatusShipped ||
		order.Status == domain.OrderStatusDelivered {
		tx.Rollback()
		return ErrOrderAlreadyProcessed
	}

	item, err := s.orderRepo.GetOrderItemForUpdate(tx, orderID, itemID)
	if err != nil {
		tx.Rollback()
		return ErrNotFound
	}

	s.logger.Debug("CancelOrderItem: item status check", "item_id", itemID, "status", item.Status)

	if !item.CanBeCancelled() {
		tx.Rollback()
		return ErrItemNotCancellable
	}

	now := time.Now().UTC()
	item.Status = domain.OrderItemStatusCancelled
	item.CancellationReason = &reason
	item.CancelledAt = &now

	if err := s.orderRepo.UpdateOrderItemTx(tx, item); err != nil {
		tx.Rollback()
		return err
	}

	product, err := s.productRead.GetByIDForUpdate(tx, item.ProductID)
	if err != nil {
		tx.Rollback()
		return err
	}

	product.Stock += int(item.Quantity)

	if err := s.productWrite.UpdateTx(tx, product); err != nil {
		tx.Rollback()
		return err
	}

	items, err := s.orderRepo.GetOrderItemsTx(tx, orderID)
	if err != nil {
		tx.Rollback()
		return err
	}

	allCancelled := true
	newTotal := 0.0

	for _, it := range items {
		if it.Status != domain.OrderItemStatusCancelled {
			allCancelled = false
			newTotal += it.Subtotal
		}
	}

	order.FinalTotal = newTotal

	if allCancelled {
		order.Status = domain.OrderStatusCancelled
	}

	if err := s.orderRepo.UpdateTx(tx, order); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *OrderService) ListOrderItems(
	orderID uint,
	userID uint,
) ([]domain.OrderItem, error) {

	order, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, ErrForbidden
	}

	return s.orderRepo.GetOrderItems(orderID)
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

		// ✅ FIXED: Get price from product_prices table
		var unitPrice float64
		var finalPrice float64

		if len(product.Prices) > 0 {
			price, err := product.GetPriceByType("H")
			if err != nil {
				if len(product.Prices) > 0 {
					unitPrice = product.Prices[0].Price
				} else {
					tx.Rollback()
					return nil, ErrPriceNotFound
				}
			} else {
				unitPrice = price
			}

			finalPrice = product.CalculatePrice("H", now)
			if finalPrice == 0 {
				finalPrice = unitPrice
			}
		} else {
			tx.Rollback()
			return nil, ErrPriceNotFound
		}

		discount := unitPrice - finalPrice
		subtotal := finalPrice * float64(req.Quantity)

		orderItems = append(orderItems, domain.OrderItem{
			ProductID:      product.ID,
			Quantity:       req.Quantity,
			Price:          unitPrice,
			DiscountAmount: discount,
			FinalPrice:     finalPrice,
			Subtotal:       subtotal,
			Status:         domain.OrderItemStatusPending,
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

	fullOrder, err := s.orderRepo.GetByIDWithAssociations(order.ID)
	if err != nil {
		return nil, err
	}

	return fullOrder, nil
}

func (s *OrderService) GetOrderByID(orderID uint) (*domain.Order, error) {
	return s.orderRepo.GetByID(orderID) // already preloads ShippingAddress ✅
}
