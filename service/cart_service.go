package service

import (
	"backend/internal/domain"
	"backend/repository"
	"errors"
)

type CartService struct {
	cartRepo    repository.CartRepositoryInterface
	productRepo repository.ProductRepositoryInterface
}

func NewCartService(
	cartRepo repository.CartRepositoryInterface,
	productRepo repository.ProductRepositoryInterface,
) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *CartService) AddItem(userID, productID uint, qty uint) error {
	if qty == 0 {
		return errors.New("quantity must be greater than zero")
	}

	if _, err := s.productRepo.GetByID(productID); err != nil {
		return errors.New("product does not exist")
	}

	cart, err := s.cartRepo.GetorCreateCart(userID)
	if err != nil {
		return err
	}

	for i := range cart.Items {
		if cart.Items[i].ProductID == productID {
			cart.Items[i].Quantity += qty
			return s.cartRepo.Save(&cart.Items[i])
		}
	}

	item := domain.CartItem{
		CartID:    cart.ID,
		ProductID: productID,
		Quantity:  qty,
	}
	return s.cartRepo.Save(&item)
}

func (s *CartService) GetCart(userID uint) (*domain.Cart, error) {
	return s.cartRepo.GetorCreateCart(userID)
}

func (s *CartService) UpdateItem(userID, itemID uint, qty uint) error {
	if qty == 0 {
		return errors.New("quantity must be greater than zero")
	}

	cart, err := s.cartRepo.GetorCreateCart(userID)
	if err != nil {
		return err
	}

	item, err := s.cartRepo.FindItem(cart.ID, itemID)
	if err != nil {
		return errors.New("cart item not found")
	}

	item.Quantity = qty
	return s.cartRepo.Save(item)
}

func (s *CartService) RemoveItem(userID, itemID uint) error {
	cart, err := s.cartRepo.GetorCreateCart(userID)
	if err != nil {
		return err
	}

	item, err := s.cartRepo.FindItem(cart.ID, itemID)
	if err != nil {
		return errors.New("cart item not found")
	}

	return s.cartRepo.Delete(item)
}
