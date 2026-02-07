package domain

import "time"

type Product struct {
    ID          uint    `gorm:"primaryKey"`
    Name        string   `gorm:"not null"`
    Description string   
    Price       float64  `gorm:"not null"`
    Stock       int      `gorm:"not null"`
    CategoryID  uint     `gorm:"not null"`
    Category    Category `gorm:"foreignKey:CategoryID;reference:ID"`
    IsActive    bool     `gorm:"default:true"`
   
	DiscountPercent *float64  `gorm:"type:decimal(5,2);default:null"` // e.g., 10.5 for 10.5%
    OfferStart      *time.Time
    OfferEnd        *time.Time
	
	CreatedAt   time.Time
    UpdatedAt   time.Time
}


type Category struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"unique;not null"`
}
