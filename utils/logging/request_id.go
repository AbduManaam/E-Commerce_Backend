package logging

import (
	"crypto/rand"
	"encoding/hex"
)

// Context key constants used across middleware for request tracing.
const (
	RequestIDKey = "requestID"
	UserIDKey    = "userID"
	RoleKey      = "role"
)

// GenerateRequestID returns a short, unique 8-character hex string
// suitable for tracing a request across log entries.
func GenerateRequestID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Fallback — should never happen in practice
		return "00000000"
	}
	return hex.EncodeToString(b)
}
