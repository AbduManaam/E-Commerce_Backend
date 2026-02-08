package dto

type CreateProductRequest struct {
    Name            string   `json:"name" validate:"required,min=3,max=100"`
    Description     string   `json:"description" validate:"required,min=10,max=500"`
    Price           float64  `json:"price" validate:"required,gt=0"`
    Stock           int      `json:"stock" validate:"required,gte=0"`
    CategoryID      uint     `json:"category_id" validate:"required"`
    
    DiscountPercent *float64 `json:"discount_percent" validate:"omitempty,gt=0,lt=100"`
    OfferStart      *string  `json:"offer_start" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"` 
    OfferEnd        *string  `json:"offer_end" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

