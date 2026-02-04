package repository

import (
	"backend/internal/domain"
	"log/slog"

	"gorm.io/gorm"
)

type CartRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewCartRepository(
	db *gorm.DB,
	logger *slog.Logger,
) *CartRepository {
	return &CartRepository{
		db:     db,
		logger: logger,
	}
}

// ------------------------------------------------------------

func (r *CartRepository) GetorCreateCart(userID uint) (*domain.Cart, error) {
	var cart domain.Cart

	err := r.db.
		Where("user_id = ?", userID).
		Preload("Items").
		First(&cart).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Info(
				"cart not found, creating new cart",
				"user_id", userID,
			)

			cart = domain.Cart{UserID: userID}
			if err := r.db.Create(&cart).Error; err != nil {
				r.logger.Error(
					"failed to create cart",
					"user_id", userID,
					"err", err,
				)
				return nil, err
			}

			return &cart, nil
		}

		r.logger.Error(
			"failed to get cart",
			"user_id", userID,
			"err", err,
		)
		return nil, err
	}

	r.logger.Info(
		"cart retrieved",
		"user_id", userID,
		"cart_id", cart.ID,
	)
	return &cart, nil
}

func (r *CartRepository) FindItem(cartID, itemID uint) (*domain.CartItem, error) {
	var cartItem domain.CartItem

	err := r.db.
		Where("cart_id = ? AND id = ?", cartID, itemID).
		First(&cartItem).Error

	if err != nil {
		r.logger.Error(
			"failed to find cart item",
			"cart_id", cartID,
			"item_id", itemID,
			"err", err,
		)
		return nil, err
	}

	r.logger.Info(
		"cart item found",
		"cart_id", cartID,
		"item_id", itemID,
	)
	return &cartItem, nil
}

func (r *CartRepository) Save(item *domain.CartItem) error {
	if err := r.db.Save(item).Error; err != nil {
		r.logger.Error(
			"failed to save cart item",
			"cart_id", item.CartID,
			"item_id", item.ID,
			"err", err,
		)
		return err
	}

	r.logger.Info(
		"cart item saved",
		"cart_id", item.CartID,
		"item_id", item.ID,
	)
	return nil
}

func (r *CartRepository) Delete(item *domain.CartItem) error {
	if err := r.db.Delete(item).Error; err != nil {
		r.logger.Error(
			"failed to delete cart item",
			"cart_id", item.CartID,
			"item_id", item.ID,
			"err", err,
		)
		return err
	}

	r.logger.Info(
		"cart item deleted",
		"cart_id", item.CartID,
		"item_id", item.ID,
	)
	return nil
}
