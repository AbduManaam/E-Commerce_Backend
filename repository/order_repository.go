package repository

import (
	"backend/internal/domain"
	"errors"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type orderRepository struct {
	db          *gorm.DB
	logger      *slog.Logger
	paymentRepo PaymentRepository
}

func NewOrderRepository(
	db *gorm.DB,
	logger *slog.Logger,
	paymentRepo PaymentRepository,
) OrderRepository {
	return &orderRepository{
		db:          db,
		logger:      logger,
		paymentRepo: paymentRepo,
	}
}

func (r *orderRepository) Begin() *gorm.DB {
	return r.db.Begin()
}

func (r *orderRepository) CreateTx(tx *gorm.DB, order *domain.Order) error {
	return tx.Create(order).Error
}

func (r *orderRepository) Create(order *domain.Order) error {
	r.logger.Info("creating order", "user_id", order.UserID)

	err := r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(order).Error
	})

	if err != nil {
		r.logger.Error("order create failed", "user_id", order.UserID, "err", err)
		return err
	}

	r.logger.Info("order created", "order_id", order.ID, "user_id", order.UserID)
	return nil
}

func (r *orderRepository) GetByID(id uint) (*domain.Order, error) {
	var order domain.Order

	err := r.db.
		Preload("Items.Product.Category").
		 Preload("Items.Product.Images").
		Preload("ShippingAddress").
		First(&order, id).Error

	if err != nil {
		r.logger.Error("order get by id failed", "order_id", id, "err", err)
		return nil, err
	}

	return &order, nil
}

func (r *orderRepository) GetByUserID(userID uint) ([]*domain.Order, error) {
	var orders []*domain.Order

	err := r.db.
		Where("user_id = ?", userID).
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.Product.Images").
		Preload("ShippingAddress").
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		r.logger.Error("get orders by user id failed", "user_id", userID, "err", err)
		return nil, err
	}

	r.logger.Info("orders fetched by user id", "user_id", userID, "count", len(orders))
	return orders, nil
}

func (r *orderRepository) GetOrdersByUserID(userID uint) ([]domain.Order, error) {
	if userID == 0 {
		r.logger.Error("invalid user id in GetOrdersByUserID")
		return nil, errors.New("invalid user id")
	}

	var orders []domain.Order

	err := r.db.
		Where("user_id = ?", userID).
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.Product.Images").
		Preload("ShippingAddress").
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		r.logger.Error("get orders by user id failed", "user_id", userID, "err", err)
		return nil, err
	}

	r.logger.Info("orders fetched by user id", "user_id", userID, "count", len(orders))
	return orders, nil
}

func (r *orderRepository) ListAll() ([]domain.Order, error) {
	var orders []domain.Order

	err := r.db.
		Preload("Items.Product.Images").
		Preload("Items.Product.Category").
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		r.logger.Error("list all orders failed", "err", err)
		return nil, err
	}

	r.logger.Info("all orders listed", "count", len(orders))
	return orders, nil
}

func (r *orderRepository) UpdateStatus(orderID uint, status domain.OrderStatus) error {
    if orderID == 0 {
        r.logger.Error("invalid order id in UpdateStatus")
        return errors.New("invalid order id")
    }

    // Fetch order first to check payment method
    var order domain.Order
    if err := r.db.First(&order, orderID).Error; err != nil {
        r.logger.Error("UpdateStatus: order not found", "order_id", orderID, "err", err)
        return err
    }

    r.logger.Info("fetched order for status update",
        "order_id", orderID,
        "payment_method", order.PaymentMethod,
        "current_payment_status", order.PaymentStatus,
        "new_status", status,
    )

    updates := map[string]interface{}{
        "status": status,
    }

    isCOD := strings.EqualFold(string(order.PaymentMethod), "cod")
    isRazorpay := strings.EqualFold(string(order.PaymentMethod), "razorpay")

    switch {
    case status == domain.OrderStatusDelivered:
        // Razorpay & COD: delivered → paid
        now := time.Now()
        updates["payment_status"] = "paid"
        updates["paid_at"] = &now

    case isRazorpay && status == domain.OrderStatusCancelled:
        // Razorpay cancelled: if payment was completed → refunded; if still pending → cancelled
        paymentStatus := string(domain.PaymentStatusCancelled)
        if payment, err := r.paymentRepo.GetByOrderID(orderID); err == nil && payment.Status == domain.PaymentStatusPaid {
            paymentStatus = string(domain.PaymentStatusRefunded)
        }
        updates["payment_status"] = paymentStatus
        updates["paid_at"] = nil

    case isRazorpay && (status == domain.OrderStatusPending || status == domain.OrderStatusShipped || status == domain.OrderStatusConfirmed):
        // Razorpay: Pending/Shipped/Confirmed → payment status stays pending
        updates["payment_status"] = "pending"
        updates["paid_at"] = nil

    case isCOD && status == domain.OrderStatusCancelled:
        // COD cancelled → payment status = cancelled (per business rules)
        updates["payment_status"] = string(domain.PaymentStatusCancelled)
        updates["paid_at"] = nil

    case isCOD && (status == domain.OrderStatusPending || status == domain.OrderStatusShipped || status == domain.OrderStatusConfirmed):
        // COD: Pending/Shipped/Confirmed → pending
        updates["payment_status"] = "pending"
        updates["paid_at"] = nil
    }

    err := r.db.
        Model(&domain.Order{}).
        Where("id = ?", orderID).
        Updates(updates).
        Error

    if err != nil {
        r.logger.Error("update order status failed",
            "order_id", orderID,
            "status", status,
            "err", err,
        )
        return err
    }

    r.logger.Info("order status updated successfully",
        "order_id", orderID,
        "status", status,
        "payment_method", order.PaymentMethod,
        "isCOD", isCOD,
    )
    return nil
}






func (r *orderRepository) Delete(id uint) error {
	if id == 0 {
		r.logger.Error("invalid order id in Delete")
		return errors.New("invalid order id")
	}

	err := r.db.Delete(&domain.Order{}, id).Error
	if err != nil {
		r.logger.Error("order delete failed", "order_id", id, "err", err)
		return err
	}

	r.logger.Info("order deleted", "order_id", id)
	return nil
}

func (r *orderRepository) GetOrdersByUserIDPaginated(userID uint, offset int, limit int) ([]domain.Order, error) {
	if userID == 0 {
		r.logger.Error("invalid user id in GetOrdersByUserIDPaginated")
		return nil, errors.New("invalid user id")
	}

	var orders []domain.Order

	err := r.db.
		Where("user_id = ?", userID).
		Preload("Items").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders).Error

	if err != nil {
		r.logger.Error("paginated orders fetch failed", "user_id", userID, "offset", offset, "limit", limit, "err", err)
		return nil, err
	}

	return orders, nil
}

func (r *orderRepository) CountOrdersByUserID(userID uint) (int64, error) {
	if userID == 0 {
		r.logger.Error("invalid user id in CountOrdersByUserID")
		return 0, errors.New("invalid user id")
	}

	var count int64

	err := r.db.
		Model(&domain.Order{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	if err != nil {
		r.logger.Error("count orders by user id failed", "user_id", userID, "err", err)
		return 0, err
	}

	return count, nil
}

func (r *orderRepository) GetOrderItem(orderID, itemID uint) (*domain.OrderItem, error) {
	var item domain.OrderItem
	err := r.db.Where("order_id = ? AND id = ?", orderID, itemID).First(&item).Error
	return &item, err
}

func (r *orderRepository) UpdateOrderItem(item *domain.OrderItem) error {
	return r.db.Save(item).Error
}

func (r *orderRepository) Update(order *domain.Order) error {
	return r.db.Save(order).Error
}

func (r *orderRepository) UpdateOrderItemTx(tx *gorm.DB, item *domain.OrderItem) error {
	return tx.Save(item).Error
}

func (r *orderRepository) UpdateTx(tx *gorm.DB, order *domain.Order) error {
	return tx.Save(order).Error
}

func (r *orderRepository) CreateOrder(order *domain.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) GetByIDWithAssociations(id uint) (*domain.Order, error) {
	var order domain.Order
	err := r.db.Preload("Items.Product").
		Preload("ShippingAddress").
		First(&order, id).Error
	return &order, err
}

func (r *orderRepository) GetOrderItems(orderID uint) ([]domain.OrderItem, error) {
	var items []domain.OrderItem
	err := r.db.
		Where("order_id = ?", orderID).
		Preload("Product").
		Find(&items).Error
	return items, err
}

func (r *orderRepository) GetOrderByID(orderID uint) (*domain.Order, error) {
	var order domain.Order
	err := r.db.
		Preload("Items").
		Preload("Items.Product").
		Preload("ShippingAddress").
		First(&order, orderID).Error
	return &order, err
}

func (r *orderRepository) UpdateOrder(order *domain.Order) error {
	return r.db.Save(order).Error
}

func (r *orderRepository) WithTransaction(fn func(repo OrderRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &orderRepository{db: tx}
		return fn(txRepo)
	})
}

func (r *orderRepository) GetByIDForUpdate(tx *gorm.DB, id uint) (*domain.Order, error) {
	var order domain.Order
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&order, id).Error
	return &order, err
}

func (r *orderRepository) GetOrderItemForUpdate(tx *gorm.DB, orderID, itemID uint) (*domain.OrderItem, error) {
	var item domain.OrderItem
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND order_id = ?", itemID, orderID).
		First(&item).Error
	return &item, err
}

func (r *orderRepository) GetOrderItemsTx(tx *gorm.DB, orderID uint) ([]domain.OrderItem, error) {
	var items []domain.OrderItem
	err := tx.
		Where("order_id = ?", orderID).
		Find(&items).Error
	return items, err
}

func (r *orderRepository) GetByIDWithItems(orderID uint) (*domain.Order, error) {
	var order domain.Order
	err := r.db.Preload("Items").First(&order, orderID).Error
	return &order, err
}

func (r *orderRepository) SaveOrderWithItems(order *domain.Order) error {
	return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(order).Error
}