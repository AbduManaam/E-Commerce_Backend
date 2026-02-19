package repository

import (
	"backend/internal/domain"
	"log/slog"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type cartRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewCartRepository(
	db *gorm.DB,
	logger *slog.Logger,
) CartRepositoryInterface {
	return &cartRepository{
		db:     db,
		logger: logger,
	}
}

// ------------------------------------------------------------

func (r *cartRepository) GetorCreateCart(userID uint) (*domain.Cart, error) {
	var cart domain.Cart

	err := r.db.
		Where("user_id = ?", userID).
		Preload("Items.Product.Category").
		Preload("Items.Product.Images").    
		Preload("Items.Product.Prices"). 
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

func (r *cartRepository) FindItem(cartID, itemID uint) (*domain.CartItem, error) {
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

func (r *cartRepository) Save(item *domain.CartItem) error {
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

func (r *cartRepository) Delete(item *domain.CartItem) error {
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

func (r *cartRepository) GetForUpdate(
	tx *gorm.DB,
	userID uint,
) (*domain.Cart, error) {

	var cart domain.Cart

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		Preload("Items").
		First(&cart).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Info(
				"cart not found for update, treating as empty cart",
				"user_id", userID,
			)

			// IMPORTANT: return empty cart, not error
			return &domain.Cart{
				UserID: userID,
				Items:  []domain.CartItem{},
			}, nil
		}

		r.logger.Error(
			"cart get for update failed",
			"user_id", userID,
			"err", err,
		)
		return nil, err
	}

	return &cart, nil
}


func (r *cartRepository) ClearTx(
	tx *gorm.DB,
	userID uint,
) error {

	if err := tx.
		Where("cart_id IN (SELECT id FROM carts WHERE user_id = ?)", userID).
		Delete(&domain.CartItem{}).
		Error; err != nil {

		r.logger.Error(
			"cart clear tx failed",
			"user_id", userID,
			"err", err,
		)
		return err
	}

	return nil
}
