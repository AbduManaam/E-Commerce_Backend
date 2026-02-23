package middleware

import (
	"backend/utils/logging"
	"fmt"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
)

func RecoveryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				requestID, _ := c.Locals(logging.RequestIDKey).(string)
				stack := string(debug.Stack())

				logging.LogError("panic recovered",
					"request_id", requestID,
					"error", fmt.Sprintf("%v", r),
					"path", c.Path(),
					"method", c.Method(),
					"stack", stack,
				)

				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
		}()
		return c.Next()
	}
}
