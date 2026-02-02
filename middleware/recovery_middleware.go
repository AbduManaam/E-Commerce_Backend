package middleware

import (
	"backend/utils/logging"

	"github.com/gofiber/fiber/v2"
)

func RecoveryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {

				var err error
				switch x := r.(type) {
				case string:
					err = fiber.NewError(fiber.StatusInternalServerError, x)
				case error:
					err = x
				default:
					err = fiber.NewError(fiber.StatusInternalServerError, "unknown panic")
				}

				logging.LogWarn("panic recovered", c, err)
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
		}()
		return c.Next()
	}
}
