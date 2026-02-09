package service

import (
	"backend/internal/domain"
	"backend/repository"
)

type AddressService struct {
	addressRepo repository.AddressRepository
}

func NewAddressService(addressRepo repository.AddressRepository) *AddressService {
	return &AddressService{addressRepo: addressRepo}
}

func (s *AddressService) Create(userID uint, address *domain.Address) (*domain.Address, error) {
	address.UserID = userID

	if err := s.addressRepo.Create(address); err != nil {
		return nil, err
	}

	return address, nil
}

func (s *AddressService) List(userID uint) ([]domain.Address, error) {
	return s.addressRepo.ListByUser(userID)
}

func (s *AddressService) UnsetOtherDefaults(userID, addressID uint) error {
	return s.addressRepo.UnsetDefaultExcept(userID, addressID)
}
