package service

import (
	"backend/internal/domain"
	"backend/repository"
	"errors"
	"log"
)

type CartService struct {
	cartRepo    repository.CartRepositoryInterface
	productRepo repository.ProductRepositoryInterface
	logger      *log.Logger
}

func NewCartService(
	cartRepo repository.CartRepositoryInterface,
	productRepo repository.ProductRepositoryInterface,
	logger *log.Logger,
) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
		logger:      logger,
	}
}

func (s *CartService) AddItem(userID, productID uint, qty uint) error {
	if qty == 0 {
		s.logger.Printf("AddItem failed: qty=0 userID=%d productID=%d", userID, productID)
		return errors.New("quantity must be greater than zero")
	}

	if _, err := s.productRepo.GetByID(productID); err != nil {
		s.logger.Printf("AddItem failed: product not found productID=%d err=%v", productID, err)
		return errors.New("product does not exist")
	}

	cart, err := s.cartRepo.GetorCreateCart(userID)
	if err != nil {
		s.logger.Printf("AddItem failed: get/create cart userID=%d err=%v", userID, err)
		return err
	}

	for i := range cart.Items {
		if cart.Items[i].ProductID == productID {
			cart.Items[i].Quantity += qty
			if err := s.cartRepo.Save(&cart.Items[i]); err != nil {
				s.logger.Printf(
					"AddItem failed: update quantity userID=%d productID=%d err=%v",
					userID, productID, err,
				)
				return err
			}
			return nil
		}
	}

	item := domain.CartItem{
		CartID:    cart.ID,
		ProductID: productID,
		Quantity:  qty,
	}

	if err := s.cartRepo.Save(&item); err != nil {
		s.logger.Printf(
			"AddItem failed: save new item userID=%d productID=%d err=%v",
			userID, productID, err,
		)
		return err
	}

	return nil
}

func (s *CartService) GetCart(userID uint) (*domain.Cart, error) {
	cart, err := s.cartRepo.GetorCreateCart(userID)
	if err != nil {
		s.logger.Printf("GetCart failed: userID=%d err=%v", userID, err)
		return nil, err
	}
	return cart, nil
}

func (s *CartService) UpdateItem(userID, itemID uint, qty uint) error {
	if qty == 0 {
		s.logger.Printf("UpdateItem failed: qty=0 userID=%d itemID=%d", userID, itemID)
		return errors.New("quantity must be greater than zero")
	}

	cart, err := s.cartRepo.GetorCreateCart(userID)
	if err != nil {
		s.logger.Printf("UpdateItem failed: get cart userID=%d err=%v", userID, err)
		return err
	}

	item, err := s.cartRepo.FindItem(cart.ID, itemID)
	if err != nil {
		s.logger.Printf(
			"UpdateItem failed: item not found userID=%d itemID=%d err=%v",
			userID, itemID, err,
		)
		return errors.New("cart item not found")
	}

	item.Quantity = qty
	if err := s.cartRepo.Save(item); err != nil {
		s.logger.Printf(
			"UpdateItem failed: save item userID=%d itemID=%d err=%v",
			userID, itemID, err,
		)
		return err
	}

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
