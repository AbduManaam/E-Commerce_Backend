package repository

import "backend/internal/domain"

type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id uint) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	Update(user *domain.User) error
	UpdatePassword(id uint, newPassword string) error
	Delete(id uint) error
}

type ProductRepository interface {
	Create(product *domain.Product) error
	GetByID(id uint) (*domain.Product, error)
	List() ([]*domain.Product, error)
	Update(product *domain.Product) error
	Delete(id uint) error
}


type OrderRepository interface {
	Create(order *domain.Order) error
	GetByID(id uint) (*domain.Order, error)
	GetOrdersByUserID(userID uint) ([]domain.Order, error)
	UpdateStatus(orderID uint, status domain.OrderStatus) error
	ListAll() ([]domain.Order, error)
	Delete(id uint) error
}

type WishlistRepositoryInterface interface {
	Add(item *domain.WishlistItem) error
	Remove(userID, productID uint) error
	GetByUserID(userID uint) ([]domain.WishlistItem, error)
}

type ProductRepositoryInterface interface {
	GetByID(productID uint) (*domain.Product, error)
	Create(p *domain.Product) error
	List() ([]*domain.Product, error)
	Update(p *domain.Product) error
	Delete(id uint) error
}
type CartRepositoryInterface interface {
	GetorCreateCart(userID uint) (*domain.Cart, error)
	FindItem(CartID, ItemID uint) (*domain.CartItem, error)
	Save(item *domain.CartItem) error
	Delete(item *domain.CartItem) error
}
