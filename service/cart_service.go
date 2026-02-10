package service

import (
	"backend/internal/domain"
	"backend/repository"
	"errors"
	"fmt"
	"log"
	"time"
)

type CartService struct {
	cartRepo    repository.CartRepositoryInterface
	productReader repository.ProductReader
	productWriter repository.ProductWriter
	logger      *log.Logger
}

func NewCartService(
	cartRepo repository.CartRepositoryInterface,
	productReader repository.ProductReader,
	productWriter repository.ProductWriter,
	logger *log.Logger,
) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productReader: productReader,
		productWriter: productWriter,
		logger:      logger,
	}
}
func (s *CartService) AddItem(userID, productID uint, qty uint) error {
	// Quantity validation
	if qty == 0 || qty > domain.MaxQuantityPerProduct {
		s.logger.Printf(
			"AddItem failed: invalid quantity userID=%d productID=%d qty=%d",
			userID, productID, qty,
		)
		return &ServiceError{
			Code: "INVALID_QUANTITY",
			Msg:  fmt.Sprintf("Quantity must be between 1 and %d", domain.MaxQuantityPerProduct),
		}
	}

	// Fetch product
	product, err := s.productReader.GetByID(productID)
	if err != nil || product == nil {
		s.logger.Printf(
			"AddItem failed: product not found userID=%d productID=%d err=%v",
			userID, productID, err,
		)
		return &ServiceError{
			Code: "PRODUCT_NOT_FOUND",
			Msg:  "Product does not exist",
		}
	}

	// Availability check
	if !product.IsActive || product.Stock <= 0 {
		s.logger.Printf(
			"AddItem failed: product unavailable userID=%d productID=%d active=%v stock=%d",
			userID, productID, product.IsActive, product.Stock,
		)
		return &ServiceError{
			Code: "PRODUCT_UNAVAILABLE",
			Msg:  "This product is currently unavailable",
		}
	}

	// Stock check
	if uint(product.Stock) < qty {
		s.logger.Printf(
			"AddItem failed: insufficient stock userID=%d productID=%d requested=%d available=%d",
			userID, productID, qty, product.Stock,
		)
		return &ServiceError{
			Code: "INSUFFICIENT_STOCK",
			Msg:  fmt.Sprintf("Only %d items available", product.Stock),
		}
	}

	// Get or create cart
	cart, err := s.cartRepo.GetorCreateCart(userID)
	if err != nil {
		s.logger.Printf(
			"AddItem failed: cart retrieval error userID=%d err=%v",
			userID, err,
		)
		return err
	}

	// Cart size limit
	if len(cart.Items) >= domain.MaxCartItems {
		s.logger.Printf(
			"AddItem failed: cart limit reached userID=%d items=%d",
			userID, len(cart.Items),
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
				s.logger.Printf(
					"AddItem failed: per-product quantity limit userID=%d productID=%d qty=%d",
					userID, productID, newQty,
				)
				return &ServiceError{
					Code: "QUANTITY_LIMIT_REACHED",
					Msg:  fmt.Sprintf("Maximum %d units per product", domain.MaxQuantityPerProduct),
				}
			}

			if uint(product.Stock) < newQty {
				s.logger.Printf(
					"AddItem failed: insufficient stock on update userID=%d productID=%d requested=%d available=%d",
					userID, productID, newQty, product.Stock,
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
		s.logger.Printf("GetCart failed: userID=%d err=%v", userID, err)
		return nil, err
	}
	now := time.Now()
	for i := range cart.Items {
		if cart.Items[i].Product.ID != 0 {
			cart.Items[i].Product.FinalPrice = cart.Items[i].Product.CalculatePrice(now)
		}
	}
	return cart, nil
}



func (s *CartService) UpdateItem(userID, itemID uint, qty uint) error {
	// Validate quantity
	if qty == 0 || qty > domain.MaxQuantityPerProduct {
		s.logger.Printf(
			"UpdateItem failed: invalid quantity userID=%d itemID=%d qty=%d",
			userID, itemID, qty,
		)
		return &ServiceError{
			Code: "INVALID_QUANTITY",
			Msg:  fmt.Sprintf("Quantity must be between 1 and %d", domain.MaxQuantityPerProduct),
		}
	}

	// Get or create cart
	cart, err := s.cartRepo.GetorCreateCart(userID)
	if err != nil {
		s.logger.Printf("UpdateItem failed: get cart userID=%d err=%v", userID, err)
		return err
	}

	// Find cart item
	item, err := s.cartRepo.FindItem(cart.ID, itemID)
	if err != nil {
		s.logger.Printf(
			"UpdateItem failed: item not found userID=%d itemID=%d err=%v",
			userID, itemID, err,
		)
		return &ServiceError{Code: "ITEM_NOT_FOUND", Msg: "Cart item not found"}
	}

	// Fetch product for stock check
	product, err := s.productReader.GetByID(item.ProductID)
	if err != nil || product == nil {
		s.logger.Printf(
			"UpdateItem failed: product not found userID=%d productID=%d err=%v",
			userID, item.ProductID, err,
		)
		return &ServiceError{Code: "PRODUCT_NOT_FOUND", Msg: "Product not found"}
	}

	// Check product availability
	if !product.IsActive || product.Stock <= 0 {
		s.logger.Printf(
			"UpdateItem failed: product unavailable userID=%d productID=%d active=%v stock=%d",
			userID, item.ProductID, product.IsActive, product.Stock,
		)
		return &ServiceError{Code: "PRODUCT_UNAVAILABLE", Msg: "Product unavailable"}
	}

	// Check stock
	if uint(product.Stock) < qty {
		s.logger.Printf(
			"UpdateItem failed: insufficient stock userID=%d productID=%d requested=%d available=%d",
			userID, item.ProductID, qty, product.Stock,
		)
		return &ServiceError{
			Code: "INSUFFICIENT_STOCK",
			Msg:  fmt.Sprintf("Only %d items available", product.Stock),
		}
	}

	// Update quantity and save
	item.Quantity = qty
	if err := s.cartRepo.Save(item); err != nil {
		s.logger.Printf(
			"UpdateItem failed: save item userID=%d itemID=%d err=%v",
			userID, itemID, err,
		)
		return err
	}

	s.logger.Printf(
		"UpdateItem succeeded: userID=%d itemID=%d newQty=%d",
		userID, itemID, qty,
	)
	return nil
}



func (s *CartService) RemoveItem(userID, itemID uint) error {
	cart, err := s.cartRepo.GetorCreateCart(userID)
	if err != nil {
		s.logger.Printf("RemoveItem failed: get cart userID=%d err=%v", userID, err)
		return err
	}

	item, err := s.cartRepo.FindItem(cart.ID, itemID)
	if err != nil {
		s.logger.Printf(
			"RemoveItem failed: item not found userID=%d itemID=%d err=%v",
			userID, itemID, err,
		)
		return errors.New("cart item not found")
	}

	if err := s.cartRepo.Delete(item); err != nil {
		s.logger.Printf(
			"RemoveItem failed: delete item userID=%d itemID=%d err=%v",
			userID, itemID, err,
		)
		return err
	}

	return nil
}
