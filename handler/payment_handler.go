package handler

import (
	"backend/internal/domain"
	"backend/service"

	"github.com/gofiber/fiber/v2"
)

type PaymentHandler struct {
	paymentSvc *service.PaymentService
}

func NewPaymentHandler(paymentSvc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentSvc: paymentSvc}
}

type CreatePaymentRequest struct {
	OrderID uint                 `json:"order_id"`
	Method  domain.PaymentMethod `json:"method"`
}

func (h *PaymentHandler) CreatePaymentIntent(c *fiber.Ctx) error {
	var req CreatePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	payment, secret, err := h.paymentSvc.CreatePaymentIntent(req.OrderID, req.Method)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"payment":       payment,
		"client_secret": secret,
	})
}
