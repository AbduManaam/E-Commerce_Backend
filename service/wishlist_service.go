package service

import (
	"backend/internal/domain"
	"backend/repository"
)

type WishlistService struct {
	repo        repository.WishlistRepositoryInterface
	productRepo repository.ProductRepositoryInterface
}

func NewWishlistService(
	wishlistRepo repository.WishlistRepositoryInterface,
	productRepo repository.ProductRepositoryInterface,
) *WishlistService {
	return &WishlistService{
		repo:        wishlistRepo,
		productRepo: productRepo,
	}
}

func (s *WishlistService) Add(userID, productID uint) error {
	item := &domain.WishlistItem{
		UserID:    userID,
		ProductID: productID,
	}
	return s.repo.Add(item)
}

func (s *WishlistService) Get(userID uint) ([]domain.WishlistItem, error) {
	return s.repo.GetByUserID(userID)
}

func (s *WishlistService) Remove(userID, productID uint) error {
	return s.repo.Remove(userID, productID)
}
