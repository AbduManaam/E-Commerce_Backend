package middleware

import (
	"backend/utils/logging"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

var sensitiveHeaders = map[string]bool{
	"Authorization": true,
	"Cookie":        true,
	"Set-Cookie":    true,
	"X-Api-Key":     true,
}

var skipPaths = map[string]bool{
	"/health":      true,
	"/favicon.ico": true,
	"/test-cors":   true,
}

func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip noisy endpoints
		path := c.Path()
		if skipPaths[path] {
			return c.Next()
		}

		// --- Request ID ---
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = logging.GenerateRequestID()
		}
		c.Locals(logging.RequestIDKey, requestID)
		c.Set("X-Request-ID", requestID)

		// --- Timer ---
		start := time.Now()

		// --- Process request ---
		chainErr := c.Next()

		// --- Collect response data ---
		duration := time.Since(start)
		status := c.Response().StatusCode()

		// Extract user context (set by auth middleware)
		userID := "anonymous"
		if id := c.Locals(logging.UserIDKey); id != nil {
			if uid, ok := id.(uint); ok && uid > 0 {
				userID = fmt.Sprintf("%d", uid)
			}
		}

		role := "guest"
		if r := c.Locals(logging.RoleKey); r != nil {
			if roleStr, ok := r.(string); ok && roleStr != "" {
				role = roleStr
			}
		}

		// --- Build log fields ---
		fields := []any{
			"request_id", requestID,
			"method", c.Method(),
			"path", path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"ip", c.IP(),
			"user_id", userID,
			"role", role,
			"bytes", c.Response().Header.ContentLength(),
		}

		// Include query string if present
		if qs := string(c.Request().URI().QueryString()); qs != "" {
			fields = append(fields, "query", qs)
		}

		// Add error info if present
		if chainErr != nil {
			fields = append(fields, "error", chainErr.Error())
		}

		switch {
		case chainErr != nil || status >= 500:
			logging.LogError("request completed", fields...)
		case duration > time.Second:
			logging.LogWarn("slow request", fields...)
		case status >= 400 && status != 401 && status != 403:
			logging.LogWarn("request completed", fields...)
		default:
			logging.LogInfo("request completed", fields...)
		}

		return chainErr
	}
}

func maskHeaders(headers map[string][]string) map[string]string {
	masked := make(map[string]string, len(headers))
	for k, v := range headers {
		if sensitiveHeaders[strings.Title(k)] {
			masked[k] = "[REDACTED]"
		} else if len(v) > 0 {
			masked[k] = v[0]
		}
	}
	return masked
}
