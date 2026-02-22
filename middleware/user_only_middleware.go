package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// UserOnlyMiddleware blocks admin users from user transactional endpoints.
// Admins in "view-only" mode can only perform read operations.
func UserOnlyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleAny := c.Locals("role")
		role, ok := roleAny.(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing or invalid role",
			})
		}

		if strings.ToUpper(role) == "ADMIN" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admins have read-only access in the user module. Transactional actions are not allowed.",
				"code":  "ADMIN_READ_ONLY",
			})
		}

		return c.Next()
	}
}
