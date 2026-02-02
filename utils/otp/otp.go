package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const (
	OTPLength   = 6
	OTPValidity = 5 * time.Minute
)

func Generate() string {
	max := big.NewInt(1000000) // 6-digit
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64())
}

func Expiry() time.Time {
	return time.Now().Add(OTPValidity)
}