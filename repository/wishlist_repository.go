package repository

import (
	"backend/internal/domain"
	"log/slog"

	"gorm.io/gorm"
)

type WishlistRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewWishlistRepository(db *gorm.DB, logger *slog.Logger) *WishlistRepository {
	return &WishlistRepository{
		db:     db,
		logger: logger,
	}
}

// Add adds a product to the user's wishlist
func (r *WishlistRepository) Add(item *domain.WishlistItem) error {
	if err := r.db.Create(item).Error; err != nil {
		r.logger.Error(
			"wishlist add failed",
			"user_id", item.UserID,
			"product_id", item.ProductID,
			"err", err,
		)
		return err
	}

	r.logger.Info(
		"wishlist item added",
		"user_id", item.UserID,
		"product_id", item.ProductID,
	)

	return nil
}

func (r *WishlistRepository) Remove(userID, productID uint) error {
	if err := r.db.
		Where("user_id = ? AND product_id = ?", userID, productID).
		Delete(&domain.WishlistItem{}).Error; err != nil {

		r.logger.Error(
			"wishlist remove failed",
			"user_id", userID,
			"product_id", productID,
			"err", err,
		)
		return err
	}

	r.logger.Info(
		"wishlist item removed",
		"user_id", userID,
		"product_id", productID,
	)

	return nil
}

func (r *WishlistRepository) GetByUserID(userID uint) ([]domain.WishlistItem, error) {
	var items []domain.WishlistItem

	if err := r.db.
		Where("user_id = ?", userID).
		Preload("Product").
		Find(&items).Error; err != nil {

		r.logger.Error(
			"wishlist fetch failed",
			"user_id", userID,
			"err", err,
		)
		return nil, err
	}

	r.logger.Info(
		"wishlist fetched",
		"user_id", userID,
		"count", len(items),
	)

	return items, nil
}
