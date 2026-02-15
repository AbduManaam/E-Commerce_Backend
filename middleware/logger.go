package middleware

import (
	"backend/utils/logging"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RequestLogger middleware that uses your existing logging system
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip logging for health checks and favicon to reduce noise
		path := c.Path()
		if path == "/health" || path == "/favicon.ico" || path == "/test-cors" {
			return c.Next()
		}

		// Start timer
		start := time.Now()
		
		// Get request details
		ip := c.IP()
		method := c.Method()
		userAgent := c.Get("User-Agent")
		referer := c.Get("Referer")
		
		// Extract user ID if available
		userID := "anonymous"
		if id := c.Locals("userID"); id != nil {
			if uid, ok := id.(uint); ok && uid > 0 {
				userID = string(rune(uid))
			}
		}
		
		// Extract role if available
		role := "guest"
		if r := c.Locals("role"); r != nil {
			if roleStr, ok := r.(string); ok {
				role = roleStr
			}
		}
		
		// Process the request
		chainErr := c.Next()
		
		// Calculate request duration
		duration := time.Since(start)
		
		// Get response details
		status := c.Response().StatusCode()
		contentLength := c.Response().Header.ContentLength()
		
		// Prepare log fields
		logFields := []interface{}{
			"method", method,
			"path", path,
			"status", status,
			"ip", ip,
			"duration_ms", duration.Milliseconds(),
			"bytes", contentLength,
			"user_id", userID,
			"role", role,
		}
		
		// Add optional fields if they exist
		if userAgent != "" {
			logFields = append(logFields, "user_agent", truncate(userAgent, 100))
		}
		if referer != "" {
			logFields = append(logFields, "referer", referer)
		}
		
		// Add query params for GET requests
		if method == "GET" && c.OriginalURL() != "" && strings.Contains(c.OriginalURL(), "?") {
			// Just log that there were query params, not the actual values (for privacy)
			logFields = append(logFields, "has_query", true)
		}
		
		// Log based on status code and error
		if chainErr != nil {
			// If there's an error in the chain, log it as error
			logFields = append(logFields, "chain_error", chainErr.Error())
			logging.LogWarn("request chain error", c, chainErr, logFields...)
		} else if status >= 500 {
			// Server errors
			logging.LogWarn("server error", c, nil, logFields...)
		} else if status >= 400 {
			// Client errors (except 401/403 which might be expected)
			if status != 401 && status != 403 {
				logging.LogWarn("client error", c, nil, logFields...)
			} else {
				logging.LogInfo("request completed", c, logFields...)
			}
		} else {
			// Successful requests
			logging.LogInfo("request completed", c, logFields...)
		}
		
		return chainErr
	}
}

// Helper function to truncate long strings
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func APILogger() fiber.Handler {
	return func(c *fiber.Ctx) error {

		path := c.Path()
		if !strings.HasPrefix(path, "/api/") && 
		   !strings.HasPrefix(path, "/auth/") && 
		   !strings.HasPrefix(path, "/user/") && 
		   !strings.HasPrefix(path, "/admin/") {
			return c.Next()
		}
		
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)
		
		logging.LogInfo("api request", c,
			"method", c.Method(),
			"path", path,
			"status", c.Response().StatusCode(),
			"duration_ms", duration.Milliseconds(),
			"ip", c.IP(),
		)
		
		return err
	}
}

func DebugLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		
		// Log request details
		logging.LogInfo("request started", c,
			"method", c.Method(),
			"path", c.Path(),
			"ip", c.IP(),
			"headers", c.GetReqHeaders(),
			"query", c.Queries(),
		)
		
		err := c.Next()
		duration := time.Since(start)
		
		// Log response details
		logging.LogInfo("request completed", c,
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration_ms", duration.Milliseconds(),
			"response_headers", c.GetRespHeaders(),
			"error", err,
		)
		
		return err
	}
}