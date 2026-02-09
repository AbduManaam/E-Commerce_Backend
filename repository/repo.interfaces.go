package repository

import (
	"backend/handler/dto"
	"backend/internal/domain"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id uint) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	Update(user *domain.User) error
	UpdatePassword(id uint, newPassword string) error
	Delete(id uint) error
	List(offset, limit int) ([]domain.User, error)
	Count() (int64, error)
}

type ProductRepository interface {
	Create(product *domain.Product) error
	GetByID(productID uint) (*domain.Product, error)
    List() ([]*domain.Product, error)               
    ListFiltered(q dto.ProductListQuery) ([]domain.Product, error) 
	Update(product *domain.Product) error
	Delete(id uint) error

	GetByIDForUpdate(tx *gorm.DB, id uint) (*domain.Product, error)
	UpdateTx(tx *gorm.DB, product *domain.Product) error
}


type OrderRepository interface {
	Begin() *gorm.DB

	CreateTx(tx *gorm.DB, order *domain.Order) error
	Create(order *domain.Order) error
	GetByID(id uint) (*domain.Order, error)
	GetOrdersByUserID(userID uint) ([]domain.Order, error)
	GetOrderItem(orderID, itemID uint) (*domain.OrderItem, error)
	UpdateStatus(orderID uint, status domain.OrderStatus) error
	ListAll() ([]domain.Order, error)
	Delete(id uint) error
	UpdateOrderItem(item *domain.OrderItem) error
	Update(order *domain.Order) error
	UpdateOrderItemTx(tx *gorm.DB, item *domain.OrderItem) error
    UpdateTx(tx *gorm.DB, order *domain.Order) error 
	GetOrdersByUserIDPaginated(userID uint, offset, limit int) ([]domain.Order, error)
    CountOrdersByUserID(userID uint) (int64, error)
}

type WishlistRepositoryInterface interface {
	Add(item *domain.WishlistItem) error
	Remove(userID, productID uint) error
	GetByUserID(userID uint) ([]domain.WishlistItem, error)
}

type CartRepositoryInterface interface {
	GetorCreateCart(userID uint) (*domain.Cart, error)
	FindItem(CartID, ItemID uint) (*domain.CartItem, error)
	Save(item *domain.CartItem) error
	Delete(item *domain.CartItem) error
	GetForUpdate(tx *gorm.DB, userID uint) (*domain.Cart, error)
    ClearTx(tx *gorm.DB, userID uint) error

}

type AddressRepository interface {
    Create(address *domain.Address) error
    ListByUser(userID uint) ([]domain.Address, error)
    UnsetDefaultExcept(userID, addressID uint) error
}
