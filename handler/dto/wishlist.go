package dto

type WishlistRequest struct{
	ProductID uint `json:"product_id" validate:"required"`
}