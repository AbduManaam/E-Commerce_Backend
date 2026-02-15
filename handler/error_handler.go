package handler

import (
	"backend/service"
	"errors"

	"github.com/gofiber/fiber/v2"
)

func HandleError(c *fiber.Ctx, err error) error {
	var se *service.ServiceError
	if errors.As(err, &se) {

		status := fiber.StatusBadRequest

		switch se.Code {

		// ---------- AUTH / USER ----------
		case "USER_NOT_FOUND":
			status = fiber.StatusNotFound
		case "INVALID_LOGIN":
			status = fiber.StatusUnauthorized
		case "USER_BLOCKED":
			status = fiber.StatusForbidden
		case "USER_EXISTS", "INVALID_INPUT", "PASSWORD_MISMATCH", "OTP_INVALID":
			status = fiber.StatusBadRequest

		case "ITEM_NOT_FOUND", "PRODUCT_NOT_FOUND":
			status = fiber.StatusNotFound
		case "INSUFFICIENT_STOCK":
			status = fiber.StatusConflict
		case "INVALID_QUANTITY":
			status = fiber.StatusBadRequest
		case "PRODUCT_UNAVAILABLE":
			status = fiber.StatusGone
		}

		return c.Status(status).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    se.Code,
				"message": se.Msg,
			},
		})
	}

	// ---------- UNKNOWN / SYSTEM ----------
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		},
	})
}
