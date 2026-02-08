package domain

import "time"

type WishlistItem struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"Index:not null"`
	ProductID uint `gorm:"Index:not null"`
	CreatedAt time.Time
 
	Product *Product `gorm:"foreignKey:ProductID"`
}
