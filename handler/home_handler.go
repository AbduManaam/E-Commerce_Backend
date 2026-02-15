package handler

import (
	"backend/service"
	"github.com/gofiber/fiber/v2"
)

type HomeHandler struct {
	service *service.HomeService
}

func NewHomeHandler(s *service.HomeService) *HomeHandler {
	return &HomeHandler{service: s}
}

func (h *HomeHandler) GetHome(c *fiber.Ctx) error {
	data, err := h.service.GetHomeData()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to load home data",
		})
	}

	return c.JSON(data)
}
