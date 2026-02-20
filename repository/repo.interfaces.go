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
	ListFiltered(query dto.ProductListQuery) ([]domain.Product, int64, error)
	Update(product *domain.Product) error
	Delete(id uint) error

	GetByIDForUpdate(tx *gorm.DB, id uint) (*domain.Product, error)
	UpdateTx(tx *gorm.DB, product *domain.Product) error
	GetNewArrivals(limit int) ([]*dto.Product, error)

	AddImage(productID uint, url string, publicID string, isPrimary bool) error

	GetImageByID(id uint) (*domain.ProductImage, error)
	DeleteImage(id uint) error
}

type OrderRepository interface {
	Begin() *gorm.DB

	CreateTx(tx *gorm.DB, order *domain.Order) error
	Create(order *domain.Order) error
	GetByID(id uint) (*domain.Order, error)
	GetOrdersByUserID(userID uint) ([]domain.Order, error)
	GetOrderItem(orderID, itemID uint) (*domain.OrderItem, error)
	UpdateStatus(orderID uint, status domain.OrderStatus) error // ✅ handles payment_status too
	ListAll() ([]domain.Order, error)
	Delete(id uint) error
	UpdateOrderItem(item *domain.OrderItem) error
	Update(order *domain.Order) error
	UpdateOrderItemTx(tx *gorm.DB, item *domain.OrderItem) error
	UpdateTx(tx *gorm.DB, order *domain.Order) error
	GetOrdersByUserIDPaginated(userID uint, offset, limit int) ([]domain.Order, error)
	CountOrdersByUserID(userID uint) (int64, error)
	GetByIDWithAssociations(id uint) (*domain.Order, error)

	GetOrderByID(orderID uint) (*domain.Order, error)
	GetOrderItems(orderID uint) ([]domain.OrderItem, error)
	UpdateOrder(order *domain.Order) error
	WithTransaction(fn func(repo OrderRepository) error) error

	GetByIDForUpdate(tx *gorm.DB, id uint) (*domain.Order, error)
	GetOrderItemForUpdate(tx *gorm.DB, orderID, itemID uint) (*domain.OrderItem, error)
	GetOrderItemsTx(tx *gorm.DB, orderID uint) ([]domain.OrderItem, error)

	GetByIDWithItems(orderID uint) (*domain.Order, error)
	SaveOrderWithItems(order *domain.Order) error
}

type WishlistRepositoryInterface interface {
	Add(item *domain.WishlistItem) error
	Remove(userID, productID uint) error
	GetByUserID(userID uint) ([]domain.WishlistItem, error)
	Exists(userID, productID uint) (bool, error)
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
	GetByID(userID, addressID uint) (*domain.Address, error)
	Update(address *domain.Address) error
	Delete(userID, addressID uint) error
	UnsetDefaultExcept(userID, addressID uint) error
}

type PaymentRepository interface {
	Create(payment *domain.Payment) error
	GetByID(id uint) (*domain.Payment, error)
	GetByOrderID(orderID uint) (*domain.Payment, error)
	Update(payment *domain.Payment) error
	GetByGatewayID(gatewayID string) (*domain.Payment, error)
	GetDB() *gorm.DB
	UpdateTx(tx *gorm.DB, payment *domain.Payment) error
	CreateTx(tx *gorm.DB, payment *domain.Payment) error
}

// HOME-----------------------------------

type HeroRepository interface {
	GetHero() (*dto.HeroBanner, error)
}

type FeatureRepository interface {
	GetAllFeatures() ([]*dto.Feature, error)
}

type ReviewRepository interface {
	GetReviews() ([]*dto.Review, error)
}