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
	FinalPrice float64 `gorm:"-"`
	
	CreatedAt   time.Time
    UpdatedAt   time.Time

}


type Category struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"unique;not null"`
}

func (p *Product) IsDiscountActive(now time.Time) bool {
	if p.DiscountPercent == nil || *p.DiscountPercent <= 0 {
		return false
	}
	if p.OfferStart != nil && now.Before(*p.OfferStart) {
		return false
	}
	if p.OfferEnd != nil && now.After(*p.OfferEnd) {
		return false
	}
	return true
}

func (p *Product) CalculatePrice(now time.Time) float64 {
	if !p.IsDiscountActive(now) {
		return p.Price
	}
	return p.Price * (1 - (*p.DiscountPercent / 100))
}