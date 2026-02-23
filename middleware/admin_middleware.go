package middleware

import (
	"backend/utils/logging"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID, _ := c.Locals(logging.RequestIDKey).(string)

		roleAny := c.Locals("role")
		role, ok := roleAny.(string)
		if !ok {
			logging.LogWarn("missing or invalid role",
				"request_id", requestID,
				"path", c.Path(),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing or invalid role",
			})
		}

		if strings.ToUpper(role) != "ADMIN" {
			logging.LogWarn("non-admin access attempt",
				"request_id", requestID,
				"role", role,
				"path", c.Path(),
			)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "admin access required",
			})
		}

		return c.Next()
	}
}
