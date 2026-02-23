package service

import (
	"backend/internal/domain"
	"backend/repository"
	"log/slog"
	"time"
)

type WishlistService struct {
	repo          repository.WishlistRepositoryInterface
	productReader repository.ProductReader
	logger        *slog.Logger
}

func NewWishlistService(
	wishlistRepo repository.WishlistRepositoryInterface,
	productReader repository.ProductReader,
	logger *slog.Logger,
) *WishlistService {
	return &WishlistService{
		repo:          wishlistRepo,
		productReader: productReader,
		logger:        logger,
	}
}

func (s *WishlistService) Add(userID, productID uint) error {
	if userID == 0 || productID == 0 {
		s.logger.Warn("wishlist add failed: invalid input", "user_id", userID, "product_id", productID)
		return ErrInvalidInput
	}

	// Check if already in wishlist
	exists, err := s.repo.Exists(userID, productID)
	if err != nil {
		s.logger.Error("wishlist add failed: exists check error", "user_id", userID, "product_id", productID, "error", err)
		return err
	}
	if exists {
		s.logger.Debug("wishlist add skipped: already exists", "user_id", userID, "product_id", productID)
		return ErrConflict.WithContext("Product already in wishlist")
	}

	// Check product exists
	if _, err := s.productReader.GetByID(productID); err != nil {
		s.logger.Warn("wishlist add failed: product not found", "product_id", productID, "error", err)
		return ErrProductNotFound
	}

	item := &domain.WishlistItem{
		UserID:    userID,
		ProductID: productID,
	}

	if err := s.repo.Add(item); err != nil {
		s.logger.Error("wishlist add failed: db error", "user_id", userID, "product_id", productID, "error", err)
		return err
	}

	s.logger.Info("wishlist add success", "user_id", userID, "product_id", productID)
	return nil
}

func (s *WishlistService) Get(userID uint) ([]domain.WishlistItem, error) {
	if userID == 0 {
		s.logger.Warn("wishlist get failed: invalid userID")
		return nil, ErrInvalidInput
	}

	items, err := s.repo.GetByUserID(userID)
	if err != nil {
		s.logger.Error("wishlist get failed", "user_id", userID, "error", err)
		return nil, err
	}

	// Calculate current prices for wishlist items
	now := time.Now()
	for i := range items {
		if items[i].Product != nil {
			items[i].Product.FinalPrice = items[i].Product.CalculatePrice("H", now)
		}
	}

	s.logger.Info("wishlist get success", "user_id", userID, "items_count", len(items))
	return items, nil
}

func (s *WishlistService) Remove(userID, productID uint) error {
	if userID == 0 || productID == 0 {
		s.logger.Warn("wishlist remove failed: invalid input", "user_id", userID, "product_id", productID)
		return ErrInvalidInput
	}

	if err := s.repo.Remove(userID, productID); err != nil {
		s.logger.Error("wishlist remove failed", "user_id", userID, "product_id", productID, "error", err)
		return err
	}

	s.logger.Info("wishlist remove success", "user_id", userID, "product_id", productID)
	return nil
}
