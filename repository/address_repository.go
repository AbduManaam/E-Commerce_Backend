package repository

import (
	"backend/internal/domain"
	"errors"

	"gorm.io/gorm"
)

type addressRepository struct {
	db *gorm.DB
}

func NewAddressRepository(db *gorm.DB) AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) Create(address *domain.Address) error {
	return r.db.Create(address).Error
}

func (r *addressRepository) ListByUser(userID uint) ([]domain.Address, error) {
	var addresses []domain.Address
	err := r.db.Where("user_id = ?", userID).Order("is_default DESC").Find(&addresses).Error
	return addresses, err
}

func (r *addressRepository) GetByID(userID, addressID uint) (*domain.Address, error) {
	var address domain.Address
	err := r.db.Where("id = ? AND user_id = ?", addressID, userID).First(&address).Error
	if err != nil {
		return nil, err
	}
	return &address, nil
}

func (r *addressRepository) Update(address *domain.Address) error {
	return r.db.Save(address).Error
}

func (r *addressRepository) Delete(userID, addressID uint) error {
	return r.db.Where("id = ? AND user_id = ?", addressID, userID).Delete(&domain.Address{}).Error
}

func (r *addressRepository) UnsetDefaultExcept(userID, addressID uint) error {
	return r.db.Model(&domain.Address{}).
		Where("user_id = ? AND id != ?", userID, addressID).
		Update("is_default", false).Error
}

func (r *addressRepository) GetByIDAndUser(
	addressID uint,
	userID uint,
) (*domain.Address, error) {

	var address domain.Address

	err := r.db.
		Where("id = ? AND user_id = ?", addressID, userID).
		First(&address).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("address not found or does not belong to user")
		}
		return nil, err
	}

	return &address, nil
}
