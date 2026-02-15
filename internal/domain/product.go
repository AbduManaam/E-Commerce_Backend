package domain

import (
	"errors"
	"time"
)

// PRODUCT
type Product struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"not null"`
	Description string
	Stock       int       `gorm:"not null"`
	CategoryID  uint      `gorm:"not null"`
	Category    Category  `gorm:"foreignKey:CategoryID;references:ID"`
	IsActive    bool      `gorm:"default:true"`

	DiscountPercent *float64
	OfferStart      *time.Time
	OfferEnd        *time.Time
	FinalPrice      float64 `gorm:"-"` // computed

	Prices    []ProductPrice   `gorm:"foreignKey:ProductID"`   // H/F pricing
	Images    []ProductImage   `gorm:"foreignKey:ProductID"`   // multiple images
	Variants  []ProductVariant `gorm:"foreignKey:ProductID"`   // sizes / optional price

	CreatedAt time.Time
	UpdatedAt time.Time
}

// PRODUCT PRICE (H/F)
type ProductPrice struct {
	ID        uint      `gorm:"primaryKey"`
	ProductID uint      `gorm:"not null;index"`
	Type      string    `gorm:"not null"` // "H" or "F"
	Price     float64   `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PRODUCT IMAGE
type ProductImage struct {
	ID        uint      `gorm:"primaryKey"`
	ProductID uint      `gorm:"not null;index"`
	URL       string    `gorm:"not null"`
	IsPrimary bool      `gorm:"default:false"`
	PublicID  string    `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PRODUCT VARIANT (size + optional price)
type ProductVariant struct {
	ID        uint          `gorm:"primaryKey"`
	ProductID uint          `gorm:"not null;index"`
	Size      string        // e.g., Small, Medium, Large
	PriceID   *uint
	Price     *ProductPrice `gorm:"foreignKey:PriceID"`
	Stock     int           `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CATEGORY
type Category struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

// --------------------
// PRODUCT HELPER METHODS
// --------------------

// IsDiscountActive checks if the product discount is currently active
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

// CalculatePrice computes the final price considering discount
func (p *Product) CalculatePrice(priceType string, now time.Time) float64 {
	// find the base price from Prices slice
	var basePrice float64
	for _, pr := range p.Prices {
		if pr.Type == priceType {
			basePrice = pr.Price
			break
		}
	}

	// fallback if not found
	if basePrice == 0 {
		return 0
	}

	// no discount
	if !p.IsDiscountActive(now) {
		return basePrice
	}

	// apply discount
	return basePrice * (1 - (*p.DiscountPercent / 100))
}


func (p *Product) GetPriceByType(priceType string) (float64, error) {
	for _, price := range p.Prices {
		if price.Type == priceType {
			return price.Price, nil
		}
	}
	return 0, errors.New("price not found")
}
