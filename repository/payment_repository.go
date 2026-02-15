package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *paymentRepository) Create(payment *domain.Payment) error {
	return r.db.Create(payment).Error
}

func (r *paymentRepository) GetByID(id uint) (*domain.Payment, error) {
	var payment domain.Payment
	if err := r.db.First(&payment, id).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetByOrderID(orderID uint) (*domain.Payment, error) {
	var payment domain.Payment
	if err := r.db.Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) Update(payment *domain.Payment) error {
	return r.db.Save(payment).Error
}

func (r *paymentRepository) GetByGatewayID(gatewayID string) (*domain.Payment, error) {
    var payment domain.Payment
    if err := r.db.Where("gateway_id = ?", gatewayID).First(&payment).Error; err != nil {
        return nil, err
    }
    return &payment, nil
}

func (r *paymentRepository) UpdateTx(tx *gorm.DB, payment *domain.Payment) error {
	return tx.Save(payment).Error
}
func (r *paymentRepository) CreateTx(tx *gorm.DB, payment *domain.Payment) error {
	return tx.Create(payment).Error
}