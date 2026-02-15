package handler

import (
	"backend/handler/dto"
	"backend/internal/domain"
	"backend/service"
	"backend/utils/logging"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type PaymentHandler struct {
	paymentSvc *service.PaymentService
}

func NewPaymentHandler(paymentSvc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentSvc: paymentSvc}
}

type CreatePaymentRequest struct {
	OrderID uint   `json:"order_id"`
	Method  string `json:"method"`
}

func (h *PaymentHandler) CreatePaymentIntent(c *fiber.Ctx) error {
	var req CreatePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	method := domain.PaymentMethod(strings.ToLower(strings.TrimSpace(req.Method)))

	switch method {
	case domain.PaymentMethodCOD, domain.PaymentMethodRazorpay, domain.PaymentMethodStripe:
	default:
		return c.Status(400).JSON(fiber.Map{"error": "unsupported payment method"})
	}

	payment, secret, err := h.paymentSvc.CreatePaymentIntent(req.OrderID, method)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"payment":       payment,
		"client_secret": secret,
	})
}


func (h *PaymentHandler) ConfirmPayment(c *fiber.Ctx) error {
    var req dto.ConfirmPaymentRequest
    if err := c.BodyParser(&req); err != nil {
        logging.LogWarn("invalid request body", c, err)
        return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
    }

    payment, err := h.paymentSvc.ConfirmPayment(req.PaymentID, req.Status)
    if err != nil {
        logging.LogWarn("failed to confirm payment", c, err)
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    logging.LogInfo("payment confirmed successfully", c, "payment_id", payment.GatewayID, "status", payment.Status)

    res := dto.ConfirmPaymentResponse{
        PaymentID: payment.GatewayID,
        Status:    string(payment.Status),
        PaidAt:    payment.PaidAt,
        Amount:    int64(payment.Amount),
        Message:   "Payment confirmed successfully",
    }

    return c.JSON(res)
}
