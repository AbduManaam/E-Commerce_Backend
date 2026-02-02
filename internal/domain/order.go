
package domain

import "time"

type OrderStatus string

const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusShipped   = "shipped"
	OrderStatusDelivered = "delivered"
	OrderStatusCanceled  = "canceled"
)

func IsValidOrderStatus(status OrderStatus) bool {
	switch status {
	case OrderStatusPending,
		OrderStatusPaid,
		OrderStatusShipped,
		OrderStatusDelivered,
		OrderStatusCanceled:
		return true
	default:
		return false
	}
}

type Order struct {
	ID        uint        `gorm:"primaryKey"`
	UserID    uint        `gorm:"not null"`
	Items     []OrderItem `gorm:"foreignKey:OrderID"`
	Total     float64     `gorm:"not null"`
    Status    OrderStatus `gorm:"type:varchar(20);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type OrderItem struct {
	ID        uint    `gorm:"primaryKey;autoIncrement"`
	OrderID  uint    `gorm:"not null;index"`
	ProductID uint   `gorm:"not null"`
	Quantity  uint    `gorm:"not null"`
	Price     float64 `gorm:"not null"`
}
