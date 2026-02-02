package repository

import (
	"backend/internal/domain"
	"errors"

	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}


func (r *orderRepository) Create(order *domain.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(order).Error
	})
}


func (r *orderRepository) GetByID(id uint) (*domain.Order, error) {
	var order domain.Order
	err := r.db.
		Preload("Items").
		First(&order, id).Error

	if err != nil {
		return nil, err
	}
	return &order, nil
}
//-------------------------------------------------------------------------------------------

func (r *orderRepository) GetByUserID(userID uint) ([]*domain.Order, error) {
	var orders []*domain.Order
	err := r.db.Where("user_id = ?", userID).Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetOrdersByUserID(userID uint) ([]domain.Order, error) {
    if userID == 0 {
        return nil, errors.New("invalid user id")
    }

    var orders []domain.Order
    err := r.db.
        Where("user_id = ?", userID).
        Order("created_at DESC").
        Find(&orders).Error

    if err != nil {
        return nil, err
    }

    return orders, nil
}
func (r *orderRepository) ListAll() ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.
		Preload("Items").
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	return orders, nil
}
func (r *orderRepository) UpdateStatus(orderID uint, status domain.OrderStatus) error {
	if orderID == 0 {
		return errors.New("invalid order id")
	}

	return r.db.
		Model(&domain.Order{}).
		Where("id = ?", orderID).
		Update("status", status).
		Error
}
func (r *orderRepository) Delete(id uint) error {
    if id == 0 {
        return errors.New("invalid order id")
    }
    return r.db.Delete(&domain.Order{}, id).Error
}

