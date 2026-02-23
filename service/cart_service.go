package service

import (
	"backend/internal/domain"
	"backend/repository"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type CartService struct {
	cartRepo      repository.CartRepositoryInterface
	productReader repository.ProductReader
	productWriter repository.ProductWriter
	logger        *slog.Logger
}

func NewCartService(
	cartRepo repository.CartRepositoryInterface,
	productReader repository.ProductReader,
	productWriter repository.ProductWriter,
	logger *slog.Logger,
) *CartService {
	return &CartService{
		cartRepo:      cartRepo,
		productReader: productReader,
		productWriter: productWriter,
		logger:        logger,
	}
}
func (s *CartService) AddItem(userID, productID uint, qty uint) error {
	// Quantity validation
	if qty == 0 || qty > domain.MaxQuantityPerProduct {
		s.logger.Warn("AddItem failed: invalid quantity",
			"user_id", userID, "product_id", productID, "qty", qty,
		)
		return &ServiceError{
			Code: "INVALID_QUANTITY",
			Msg:  fmt.Sprintf("Quantity must be between 1 and %d", domain.MaxQuantityPerProduct),
		}
	}

	// Fetch product
	product, err := s.productReader.GetByID(productID)
	if err != nil || product == nil {
		s.logger.Warn("AddItem failed: product not found",
			"user_id", userID, "product_id", productID, "error", err,
		)
		return &ServiceError{
			Code: "PRODUCT_NOT_FOUND",
			Msg:  "Product does not exist",
		}
	}

	// Availability check
	if !product.IsActive || product.Stock <= 0 {
		s.logger.Warn("AddItem failed: product unavailable",
			"user_id", userID, "product_id", productID, "active", product.IsActive, "stock", product.Stock,
		)
		return &ServiceError{
			Code: "PRODUCT_UNAVAILABLE",
			Msg:  "This product is currently unavailable",
		}
	}

	// Stock check
	if uint(product.Stock) < qty {
		s.logger.Warn("AddItem failed: insufficient stock",
			"user_id", userID, "product_id", productID, "requested", qty, "available", product.Stock,
		)
		return &ServiceError{
			Code: "INSUFFICIENT_STOCK",
			Msg:  fmt.Sprintf("Only %d items available", product.Stock),
		}
	}

	// Get or create cart
	cart, err := s.cartRepo.GetorCreateCart(userID)
	if err != nil {
		s.logger.Error("AddItem failed: cart retrieval error",
			"user_id", userID, "error", err,
		)
		return err
	}

	// Cart size limit
	if len(cart.Items) >= domain.MaxCartItems {
		s.logger.Warn("AddItem failed: cart limit reached",
			"user_id", userID, "items", len(cart.Items),
		)
		return &ServiceError{
			Code: "CART_LIMIT_REACHED",
			Msg:  fmt.Sprintf("Maximum %d items allowed in cart", domain.MaxCartItems),
		}
	}

	// Existing item
	for i := range cart.Items {
		if cart.Items[i].ProductID == productID {
			newQty := cart.Items[i].Quantity + qty

			if newQty > domain.MaxQuantityPerProduct {
				s.logger.Warn("AddItem failed: per-product quantity limit",
					"user_id", userID, "product_id", productID, "qty", newQty,
				)
				return &ServiceError{
					Code: "QUANTITY_LIMIT_REACHED",
					Msg:  fmt.Sprintf("Maximum %d units per product", domain.MaxQuantityPerProduct),
				}
			}

			if uint(product.Stock) < newQty {
				s.logger.Warn("AddItem failed: insufficient stock on update",
					"user_id", userID, "product_id", productID, "requested", newQty, "available", product.Stock,
				)
				return &ServiceError{
					Code: "INSUFFICIENT_STOCK",
					Msg:  fmt.Sprintf("Only %d items available", product.Stock),
				}
			}

			cart.Items[i].Quantity = newQty
			return s.cartRepo.Save(&cart.Items[i])
		}
	}

	// New item
	item := domain.CartItem{
		CartID:    cart.ID,
		ProductID: productID,
		Quantity:  qty,
	}

	return s.cartRepo.Save(&item)
}

func (s *CartService) GetCart(userID uint) (*domain.Cart, error) {
	cart, err := s.cartRepo.GetorCreateCart(userID)
	if err != nil {
		s.logger.Error("GetCart failed", "user_id", userID, "error", err)
		return nil, err
	}
	now := time.Now()
	for i := range cart.Items {
		if cart.Items[i].Product.ID != 0 {
			cart.Items[i].Product.FinalPrice = cart.Items[i].Product.CalculatePrice("H", now)
		}
	}
	return cart, nil
}

func (s *CartService) UpdateItem(userID, itemID uint, qty uint) error {
	// Validate quantity
	if qty == 0 || qty > domain.MaxQuantityPerProduct {
		s.logger.Warn("UpdateItem failed: invalid quantity",
			"user_id", userID, "item_id", itemID, "qty", qty,
		)
		return &ServiceError{
			Code: "INVALID_QUANTITY",
			Msg:  fmt.Sprintf("Quantity must be between 1 and %d", domain.MaxQuantityPerProduct),
		}
	}

	// Get or create cart
	cart, err := s.cartRepo.GetorCreateCart(userID)
	if err != nil {
		s.logger.Error("UpdateItem failed: get cart", "user_id", userID, "error", err)
		return err
	}

	// Find cart item
	item, err := s.cartRepo.FindItem(cart.ID, itemID)
	if err != nil {
		s.logger.Warn("UpdateItem failed: item not found",
			"user_id", userID, "item_id", itemID, "error", err,
		)
		return &ServiceError{Code: "ITEM_NOT_FOUND", Msg: "Cart item not found"}
	}

	// Fetch product for stock check
	product, err := s.productReader.GetByID(item.ProductID)
	if err != nil || product == nil {
		s.logger.Warn("UpdateItem failed: product not found",
			"user_id", userID, "product_id", item.ProductID, "error", err,
		)
		return &ServiceError{Code: "PRODUCT_NOT_FOUND", Msg: "Product not found"}
	}

	// Check product availability
	if !product.IsActive || product.Stock <= 0 {
		s.logger.Warn("UpdateItem failed: product unavailable",
			"user_id", userID, "product_id", item.ProductID, "active", product.IsActive, "stock", product.Stock,
		)
		return &ServiceError{Code: "PRODUCT_UNAVAILABLE", Msg: "Product unavailable"}
	}

	// Check stock
	if uint(product.Stock) < qty {
		s.logger.Warn("UpdateItem failed: insufficient stock",
			"user_id", userID, "product_id", item.ProductID, "requested", qty, "available", product.Stock,
		)
		return &ServiceError{
			Code: "INSUFFICIENT_STOCK",
			Msg:  fmt.Sprintf("Only %d items available", product.Stock),
		}
	}

	// Update quantity and save
	item.Quantity = qty
	if err := s.cartRepo.Save(item); err != nil {
		s.logger.Error("UpdateItem failed: save item",
			"user_id", userID, "item_id", itemID, "error", err,
		)
		return err
	}

	s.logger.Info("UpdateItem succeeded",
		"user_id", userID, "item_id", itemID, "new_qty", qty,
	)
	return nil
}

func (s *CartService) RemoveItem(userID, itemID uint) error {
	cart, err := s.cartRepo.GetorCreateCart(userID)
	if err != nil {
		s.logger.Error("RemoveItem failed: get cart", "user_id", userID, "error", err)
		return err
	}

	item, err := s.cartRepo.FindItem(cart.ID, itemID)
	if err != nil {
		s.logger.Warn("RemoveItem failed: item not found",
			"user_id", userID, "item_id", itemID, "error", err,
		)
		return errors.New("cart item not found")
	}

	if err := s.cartRepo.Delete(item); err != nil {
		s.logger.Error("RemoveItem failed: delete item",
			"user_id", userID, "item_id", itemID, "error", err,
		)
		return err
	}

	return nil
}
