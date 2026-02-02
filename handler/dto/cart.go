package dto

type AddToCartRequest struct{
	ProductID uint `json:"product_id" validate:"required"`
	Quantity  uint   `json:"quantity" validate:"required,min=1"`
}

type UpdateCartItems struct{
	Quantity  uint `json:"quantity" validate:"required,min=1"`
}