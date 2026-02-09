package dto

import "backend/internal/domain"

type CreateOrderRequest struct {
	AddressID     uint                 `json:"address_id" validate:"required,gt=0"`
	PaymentMethod domain.PaymentMethod `json:"payment_method" validate:"required,oneof=cod razorpay stripe paypal"`
}

type CreateDirectOrderRequest struct {
	AddressID     uint                 `json:"address_id" validate:"required,gt=0"`
	PaymentMethod domain.PaymentMethod `json:"payment_method" validate:"required,oneof=cod razorpay stripe paypal"`
	Items         []OrderItemRequest   `json:"items" validate:"required,min=1,dive"`
}

type OrderItemRequest struct {
	ProductID uint `json:"product_id" validate:"required,gt=0"`
	Quantity  uint `json:"quantity" validate:"required,gt=0"`
}