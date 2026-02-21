package middleware

import (
	"backend/utils/logging"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleAny := c.Locals("role")
		role, ok := roleAny.(string)
		if !ok {
			logging.LogWarn("missing or invalid role", c, fiber.ErrUnauthorized)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing or invalid role",
			})
		}

		if strings.ToUpper(role) != "ADMIN" {
			logging.LogWarn("non-admin access attempt", c, fiber.ErrForbidden, "role", role)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "admin access required",
			})
		}

		// Optional: successful admin log (debug/info)
		logging.LogInfo("admin access granted", c, "role", role)


		return c.Next()
	}
}
