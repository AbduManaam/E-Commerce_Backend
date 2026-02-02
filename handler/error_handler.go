package handler

import (
	"backend/service"
	"github.com/gofiber/fiber/v2"
)

func HandleError(c *fiber.Ctx, err error) error {
	if svcErr, ok := err.(*service.ServiceError); ok {
		switch svcErr.Code {
		case "USER_NOT_FOUND":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": svcErr.Msg})
		case "INVALID_LOGIN":
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": svcErr.Msg})
		case "USER_BLOCKED":
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": svcErr.Msg})
		case "USER_EXISTS", "INVALID_INPUT", "PASSWORD_MISMATCH", "OTP_INVALID":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": svcErr.Msg})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		}
	}

	// Fallback for unknown errors
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
}
