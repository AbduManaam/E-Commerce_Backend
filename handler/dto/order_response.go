package dto

import "time"

type OrderResponse struct {
	ID              uint                `json:"id"`
	Status          string              `json:"status"`
	Total           float64             `json:"total"`
	Discount        float64             `json:"discount"`
	FinalTotal      float64             `json:"final_total"`
	PaymentMethod   string              `json:"payment_method"`
	PaymentStatus   string              `json:"payment_status"`
	CreatedAt       time.Time           `json:"created_at"`
	ShippingAddress *OrderAddressDTO    `json:"shipping_address"`
	Items           []OrderItemResponse `json:"items"`
}

type OrderAddressDTO struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	City     string `json:"city"`
	State    string `json:"state"`
	Country  string `json:"country"`
	ZipCode  string `json:"zip_code"`
	Landmark string `json:"landmark"`
}

type OrderItemResponse struct {
	ID             uint            `json:"id"`
	Quantity       uint            `json:"quantity"`
	Price          float64         `json:"price"`
	DiscountAmount float64         `json:"discount_amount"`
	FinalPrice     float64         `json:"final_price"`
	Subtotal       float64         `json:"subtotal"`
	Status         string          `json:"status"`
	Product        ProductResponse `json:"product"`
}