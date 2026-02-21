package domain

import (
	"encoding/json"
	"errors"
	"time"
)

type PaymentMethod string

const (
    PaymentMethodCOD     PaymentMethod = "cod"
    PaymentMethodRazorpay PaymentMethod = "razorpay"
    PaymentMethodStripe   PaymentMethod = "stripe"
    PaymentMethodPaypal   PaymentMethod = "paypal"
)

type PaymentStatus string

const (
    PaymentStatusPending   PaymentStatus = "pending"
    PaymentStatusPaid      PaymentStatus = "paid"
    PaymentStatusFailed    PaymentStatus = "failed"
    PaymentStatusRefunded  PaymentStatus = "refunded"
    PaymentStatusCancelled PaymentStatus = "cancelled"
)

type Payment struct {
    ID            uint          `gorm:"primaryKey"`
    OrderID       uint          `gorm:"not null;index"`
    PaymentMethod PaymentMethod `gorm:"type:varchar(20);not null"`
    Amount        float64       `gorm:"not null"`
    Currency      string        `gorm:"default:'INR'"`
    Status        PaymentStatus `gorm:"type:varchar(20);not null"`

    GatewayID     string        `gorm:"type:varchar(100)"`
    GatewayData   json.RawMessage `gorm:"type:jsonb;default:'{}'"`
    FailureReason string        `gorm:"type:text"`

    PaidAt        *time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

var ErrRecordNotFound = errors.New("record not found")
