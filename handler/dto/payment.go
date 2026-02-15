package dto

import "time"

// Request body to confirm a payment
type ConfirmPaymentRequest struct {
    PaymentID string `json:"payment_id" validate:"required"`
    Method    string `json:"method" validate:"required"`    
    Status    string `json:"status" validate:"required"`     
}

// Response sent back to client
type ConfirmPaymentResponse struct {
    PaymentID string `json:"payment_id"`
    Status    string `json:"status"`
    PaidAt    *time.Time `json:"paid_at,omitempty"`
    Amount    int64  `json:"amount"`
    Message   string `json:"message"`
}
