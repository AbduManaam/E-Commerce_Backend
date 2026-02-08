package domain

import "time"

type OrderStatus string

const (
    OrderStatusPending   OrderStatus = "pending"
    OrderStatusConfirmed OrderStatus = "confirmed"
    OrderStatusPaid      OrderStatus = "paid"
    OrderStatusShipped   OrderStatus = "shipped"
    OrderStatusDelivered OrderStatus = "delivered"
    OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
    ID        uint        `gorm:"primaryKey"`
    UserID    uint        `gorm:"not null;index"`
    Items     []OrderItem `gorm:"foreignKey:OrderID"`
    Total     float64     `gorm:"not null"`
    Discount  float64     `gorm:"default:0"`
    FinalTotal float64    `gorm:"not null"`
    Status    OrderStatus `gorm:"type:varchar(20);not null;index"`
    CreatedAt time.Time
    UpdatedAt time.Time
    
    // Snapshot of shipping address
    ShippingAddressID *uint
    ShippingAddress   *OrderAddress `gorm:"foreignKey:ShippingAddressID"`
    
    // Payment info
    PaymentMethod PaymentMethod `gorm:"type:varchar(20)"`
    PaymentStatus PaymentStatus `gorm:"type:varchar(20)"`
}

type OrderItem struct {
    ID              uint    `gorm:"primaryKey"`
    OrderID         uint    `gorm:"not null;index"`
    ProductID       uint    `gorm:"not null"`
    Quantity        uint    `gorm:"not null"`
    Price           float64 `gorm:"not null"`           // Original price
    DiscountPercent float64 `gorm:"default:0"`          // Applied discount at order time
    FinalPrice      float64 `gorm:"not null"`           // Price after discount (frozen)
    Subtotal        float64 `gorm:"not null"`           // Quantity × FinalPrice
    Status          OrderItemStatus `gorm:"type:varchar(20)"`
    CancellationReason *string       `gorm:"type:text"`
    CancelledAt     *time.Time
    
    // Relationships
    Product Product `gorm:"foreignKey:ProductID"`
}

type OrderItemStatus string

const (
    OrderItemStatusPending   OrderItemStatus = "pending"
    OrderItemStatusConfirmed OrderItemStatus = "confirmed"
    OrderItemStatusCancelled OrderItemStatus = "cancelled"
    OrderItemStatusRefunded  OrderItemStatus = "refunded"
)

// Order address snapshot
type OrderAddress struct {
    ID        uint   `gorm:"primaryKey"`
    FullName  string `gorm:"not null"`
    Phone     string `gorm:"not null"`
    Address   string `gorm:"type:text;not null"`
    City      string `gorm:"not null"`
    State     string `gorm:"not null"`
    Country   string `gorm:"not null"`
    ZipCode   string `gorm:"not null"`
    Landmark  string
}