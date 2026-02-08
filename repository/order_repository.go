package repository

import (
	"backend/internal/domain"
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

type orderRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewOrderRepository(
	db *gorm.DB,
	logger *slog.Logger,
) OrderRepository {
	return &orderRepository{
		db:     db,
		logger: logger,
	}
}

// ------------------------------------------------------------

func (r *orderRepository) Begin() *gorm.DB {
	return r.db.Begin()
}

func (r *orderRepository) CreateTx(tx *gorm.DB, order *domain.Order) error {
	return tx.Create(order).Error
}



func (r *orderRepository) Create(order *domain.Order) error {
	r.logger.Info(
		"creating order",
		"user_id", order.UserID,
	)

	err := r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(order).Error
	})

	if err != nil {
		r.logger.Error(
			"order create failed",
			"user_id", order.UserID,
			"err", err,
		)
		return err
	}

	r.logger.Info(
		"order created",
		"order_id", order.ID,
		"user_id", order.UserID,
	)
	return nil
}

func (r *orderRepository) GetByID(id uint) (*domain.Order, error) {
	var order domain.Order

	err := r.db.
		Preload("Items").
		First(&order, id).Error

	if err != nil {
		r.logger.Error(
			"order get by id failed",
			"order_id", id,
			"err", err,
		)
		return nil, err
	}

	return &order, nil
}

// ------------------------------------------------------------

func (r *orderRepository) GetByUserID(userID uint) ([]*domain.Order, error) {
	var orders []*domain.Order

	err := r.db.
		Where("user_id = ?", userID).
		Find(&orders).Error

	if err != nil {
		r.logger.Error(
			"get orders by user id failed",
			"user_id", userID,
			"err", err,
		)
		return nil, err
	}

	r.logger.Info(
		"orders fetched by user id",
		"user_id", userID,
		"count", len(orders),
	)
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
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		r.logger.Error(
			"get orders by user id failed",
			"user_id", userID,
			"err", err,
		)
		return nil, err
	}

	r.logger.Info(
		"orders fetched by user id",
		"user_id", userID,
		"count", len(orders),
	)
	return orders, nil
}

func (r *orderRepository) ListAll() ([]domain.Order, error) {
	var orders []domain.Order

	err := r.db.
		Preload("Items").
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		r.logger.Error(
			"list all orders failed",
			"err", err,
		)
		return nil, err
	}

	r.logger.Info(
		"all orders listed",
		"count", len(orders),
	)
	return orders, nil
}

func (r *orderRepository) UpdateStatus(orderID uint, status domain.OrderStatus) error {
	if orderID == 0 {
		r.logger.Error("invalid order id in UpdateStatus")
		return errors.New("invalid order id")
	}

	err := r.db.
		Model(&domain.Order{}).
		Where("id = ?", orderID).
		Update("status", status).
		Error

	if err != nil {
		r.logger.Error(
			"update order status failed",
			"order_id", orderID,
			"status", status,
			"err", err,
		)
		return err
	}

	r.logger.Info(
		"order status updated",
		"order_id", orderID,
		"status", status,
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
		r.logger.Error(
			"order delete failed",
			"order_id", id,
			"err", err,
		)
		return err
	}

	r.logger.Info(
		"order deleted",
		"order_id", id,
	)
	return nil
}
