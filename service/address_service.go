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

	if address.IsDefault {
		_ = s.addressRepo.UnsetDefaultExcept(userID, address.ID)
	}

	return address, nil
}

func (s *AddressService) List(userID uint) ([]domain.Address, error) {
	return s.addressRepo.ListByUser(userID)
}

func (s *AddressService) GetByID(userID, addressID uint) (*domain.Address, error) {
	return s.addressRepo.GetByID(userID, addressID)
}

func (s *AddressService) Update(userID uint, address *domain.Address) error {
	existing, err := s.addressRepo.GetByID(userID, address.ID)
	if err != nil {
		return err
	}

	address.UserID = existing.UserID
	return s.addressRepo.Update(address)
}

func (s *AddressService) Delete(userID, addressID uint) error {
	return s.addressRepo.Delete(userID, addressID)
}

func (s *AddressService) SetDefault(userID, addressID uint) error {
	addr, err := s.addressRepo.GetByID(userID, addressID)
	if err != nil {
		return err
	}

	addr.IsDefault = true
	if err := s.addressRepo.Update(addr); err != nil {
		return err
	}

	return s.addressRepo.UnsetDefaultExcept(userID, addressID)
}
