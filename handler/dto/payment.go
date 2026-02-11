package dto

import "time"

// Request body to confirm a payment
type ConfirmPaymentRequest struct {
    PaymentID string `json:"payment_id" validate:"required"` // gateway payment ID
    Method    string `json:"method" validate:"required"`     // e.g., "razorpay", "stripe"
    Status    string `json:"status" validate:"required"`     // e.g., "success", "failed"
}

// Response sent back to client
type ConfirmPaymentResponse struct {
    PaymentID string `json:"payment_id"`
    Status    string `json:"status"`
    PaidAt    *time.Time `json:"paid_at,omitempty"`
    Amount    int64  `json:"amount"`
    Message   string `json:"message"`
}
