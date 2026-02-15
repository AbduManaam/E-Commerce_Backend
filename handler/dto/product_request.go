package dto

type PriceDTO struct {
	H *float64 `json:"H"`
	F *float64 `json:"F"`
}

type VariantDTO struct {
	Size  string   `json:"size"`
	H     *float64 `json:"H"`
	F     *float64 `json:"F"`
	Stock int      `json:"stock"`
}

type CreateProductRequest struct {
	Name        string       `json:"title"`
	Description string       `json:"description"`
	Stock       int          `json:"stock"`
	CategoryID  uint         `json:"category_id"`
	Prices      PriceDTO     `json:"price"`
	Images      []string     `json:"images"`
	Variants    []VariantDTO `json:"variants"`
	DiscountPercent *float64 `json:"discount_percent"`
	OfferStart      *string  `json:"offer_start"`
	OfferEnd        *string  `json:"offer_end"`
}
