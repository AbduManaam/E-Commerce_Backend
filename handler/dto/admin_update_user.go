package dto

type AdminUpdateUserRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50"`
	Role string `json:"role" validate:"required,oneof=USER ADMIN"`
}

type UpdateProductRequest struct {
    Name            *string  `json:"name" validate:"omitempty,min=2,max=100"`
    Description     *string  `json:"description" validate:"omitempty,max=500"`
    Price           *float64 `json:"price" validate:"omitempty,gt=0,lt=1000000"`
    Stock           *int     `json:"stock" validate:"omitempty,gte=0"`
    IsActive        *bool    `json:"is_active"`
    
    DiscountPercent *float64 `json:"discount_percent" validate:"omitempty,gt=0,lt=100"`
    OfferStart      *string  `json:"offer_start" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
    OfferEnd        *string  `json:"offer_end" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}
