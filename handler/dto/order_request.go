package dto

type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items" validate:"required,min=1,dive"`
	//dive → now validate each OrderItemRequest inside the slice
}

type OrderItemRequest struct {
	ProductID uint `json:"product_id" validate:"required,gt=0"`
	Quantity  uint  `json:"quantity" validate:"required,gt=0"`
}