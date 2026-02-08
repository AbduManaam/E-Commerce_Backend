package domain

import "time"

type Cart struct {
	ID         uint       `gorm:"primaryKey"`
	UserID     uint       `gorm:"Index;not null"`
	Items      []CartItem `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CartItem struct{

	ID uint `gorm:"primaryKey"`
	CartID uint `gorm:"Index;not null"`
	ProductID uint  `gorm:"Index;not null"`
    Quantity  uint   `gorm:"not null;check:quantity>0"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Product Product `gorm:"foreignKey:ProductID"`

}

const (
	MaxCartItems          = 20
	MaxQuantityPerProduct = 10
)
