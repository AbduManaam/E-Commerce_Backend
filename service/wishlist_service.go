package service

import (
	"backend/internal/domain"
	"backend/repository"
	"log"
	"time"
)

type WishlistService struct {
	repo        repository.WishlistRepositoryInterface
	productReader repository.ProductReader
	logger      *log.Logger
}

func NewWishlistService(
	wishlistRepo repository.WishlistRepositoryInterface,
	productReader repository.ProductReader,
	logger *log.Logger,
) *WishlistService {
	return &WishlistService{
		repo:        wishlistRepo,
		productReader: productReader,
		logger:      logger,
	}
}

func (s *WishlistService) Add(userID, productID uint) error {
	if userID == 0 || productID == 0 {
		s.logger.Printf("Wishlist Add failed: invalid input userID=%d productID=%d", userID, productID)
		return ErrInvalidInput
	}

	// Check product exists
	if _, err := s.productReader.GetByID(productID); err != nil {
		s.logger.Printf("Wishlist Add failed: product not found productID=%d err=%v", productID, err)
		return ErrProductNotFound
	}

	item := &domain.WishlistItem{
		UserID:    userID,
		ProductID: productID,
	}

	if err := s.repo.Add(item); err != nil {
		s.logger.Printf("Wishlist Add failed: db error userID=%d productID=%d err=%v", userID, productID, err)
		return err
	}

	s.logger.Printf("Wishlist Add success: userID=%d productID=%d", userID, productID)
	return nil
}




func (s *WishlistService) Get(userID uint) ([]domain.WishlistItem, error) {
	if userID == 0 {
		s.logger.Println("Wishlist Get failed: invalid userID=0")
		return nil, ErrInvalidInput
	}

	items, err := s.repo.GetByUserID(userID)
	if err != nil {
		s.logger.Printf("Wishlist Get failed: userID=%d err=%v", userID, err)
		return nil, err
	}

	// Calculate current prices for wishlist items
	now := time.Now()
    for i := range items {
        if items[i].Product != nil {
            items[i].Product.FinalPrice = items[i].Product.CalculatePrice(now)
        }
    }


	s.logger.Printf("Wishlist Get success: userID=%d itemsCount=%d", userID, len(items))
	return items, nil
}





func (s *WishlistService) Remove(userID, productID uint) error {
	if userID == 0 || productID == 0 {
		s.logger.Printf("Wishlist Remove failed: invalid input userID=%d productID=%d", userID, productID)
		return ErrInvalidInput
	}

	if err := s.repo.Remove(userID, productID); err != nil {
		s.logger.Printf("Wishlist Remove failed: userID=%d productID=%d err=%v", userID, productID, err)
		return err
	}

	s.logger.Printf("Wishlist Remove success: userID=%d productID=%d", userID, productID)
	return nil
}
