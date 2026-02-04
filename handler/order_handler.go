package handler

import (
	"strconv"

	"backend/handler/dto"
	"backend/internal/domain"
	"backend/service"
	"backend/utils/logging"
	validator "backend/utils/validation"

	"github.com/gofiber/fiber/v2"
)

type OrderHandler struct {
	orderSvc *service.OrderService
}

func NewOrderHandler(orderSvc *service.OrderService) *OrderHandler {
	return &OrderHandler{orderSvc: orderSvc}
}

func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	userIDAny := c.Locals("userID")
	userID, ok := userIDAny.(uint)
	if !ok {
		logging.LogWarn("unauthorized create order attempt", c, fiber.ErrUnauthorized)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req dto.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("invalid create order request: body parse failed", c, err, "userID", userID)
		return HandleError(c, service.ErrInvalidInput)
	}

	if err := validator.Validate.Struct(req); err != nil {
		logging.LogWarn("invalid create order request: validation failed", c, err, "userID", userID)
		return c.Status(400).JSON(fiber.Map{
			"errors": validator.FormatErrors(err),
		})
	}

	order, err := h.orderSvc.CreateOrder(userID, req)
	if err != nil {
		logging.LogWarn("create order failed: service error", c, err, "userID", userID)
		return HandleError(c, err)
	}

	logging.LogInfo("order created successfully", c, "userID", userID, "orderID", order.ID)
	return c.Status(fiber.StatusCreated).JSON(order)
}

func (h *OrderHandler) GetUserOrders(c *fiber.Ctx) error {
	userIDAny := c.Locals("userID")
	userID, ok := userIDAny.(uint)
	if !ok {
		logging.LogWarn("unauthorized get user orders attempt", c, fiber.ErrUnauthorized)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	orders, err := h.orderSvc.GetUserOrders(userID)
	if err != nil {
		logging.LogWarn("get user orders failed", c, err, "userID", userID)
		return HandleError(c, err)
	}

	logging.LogInfo("user orders retrieved successfully", c, "userID", userID, "ordersCount", len(orders))
	return c.JSON(orders)
}

func (h *OrderHandler) GetOrder(c *fiber.Ctx) error {
	userIDAny := c.Locals("userID")
	userID, ok := userIDAny.(uint)
	if !ok {
		logging.LogWarn("unauthorized get order attempt", c, fiber.ErrUnauthorized)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	idParam := c.Params("id")
	orderID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		logging.LogWarn("get order failed: invalid order ID", c, err, "userID", userID, "orderIDParam", idParam)
		return HandleError(c, service.ErrInvalidInput)
	}

	order, err := h.orderSvc.GetOrder(userID, uint(orderID))
	if err != nil {
		logging.LogWarn("get order failed: service error", c, err, "userID", userID, "orderID", orderID)
		return HandleError(c, err)
	}

	logging.LogInfo("order retrieved successfully", c, "userID", userID, "orderID", orderID)
	return c.JSON(order)
}

func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	userIDAny := c.Locals("userID")
	userID, ok := userIDAny.(uint)
	if !ok {
		logging.LogWarn("unauthorized cancel order attempt", c, fiber.ErrUnauthorized)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	idParam := c.Params("id")
	orderID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		logging.LogWarn("cancel order failed: invalid order ID", c, err, "userID", userID, "orderIDParam", idParam)
		return HandleError(c, service.ErrInvalidInput)
	}

	err = h.orderSvc.CancelOrder(userID, uint(orderID))
	if err != nil {
		logging.LogWarn("cancel order failed: service error", c, err, "userID", userID, "orderID", orderID)
		return HandleError(c, err)
	}

	logging.LogInfo("order cancelled successfully", c, "userID", userID, "orderID", orderID)
	return c.JSON(fiber.Map{"message": "Order cancelled successfully"})
}

// Admin methods
func (h *OrderHandler) ListAllOrders(c *fiber.Ctx) error {
	orders, err := h.orderSvc.ListAllOrders()
	if err != nil {
		logging.LogWarn("list all orders failed: service error", c, err)
		return HandleError(c, err)
	}

	logging.LogInfo("all orders retrieved successfully", c, "ordersCount", len(orders))
	return c.JSON(orders)
}

func (h *OrderHandler) UpdateOrderStatus(c *fiber.Ctx) error {
	idParam := c.Params("id")
	orderID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		logging.LogWarn("update order status failed: invalid order ID", c, err)
		return HandleError(c, service.ErrInvalidInput)
	}

	var req struct {
		Status string `json:"status" validate:"required,oneof=pending confirmed shipped delivered cancelled"`
	}
	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("update order status failed: body parse", c, err, "orderID", orderID)
		return HandleError(c, service.ErrInvalidInput)
	}

	if err := validator.Validate.Struct(req); err != nil {
		logging.LogWarn("update order status failed: validation error", c, err, "orderID", orderID)
		return c.Status(400).JSON(fiber.Map{"errors": validator.FormatErrors(err)})
	}

	if err := h.orderSvc.UpdateOrderStatus(uint(orderID), domain.OrderStatus(req.Status)); err != nil {
		logging.LogWarn("update order status failed: service error", c, err, "orderID", orderID)
		return HandleError(c, err)
	}

	logging.LogInfo("order status updated successfully", c, "orderID", orderID, "status", req.Status)
	return c.JSON(fiber.Map{"message": "Order status updated successfully"})
}

func (h *OrderHandler) AdminUpdateOrder(c *fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		logging.LogWarn("admin update order failed: invalid order ID", c, err)
		return HandleError(c, service.ErrInvalidInput)
	}

	var req struct {
		Status domain.OrderStatus `json:"status" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("admin update order failed: body parse", c, err, "orderID", orderID)
		return HandleError(c, service.ErrInvalidInput)
	}

	if !domain.IsValidOrderStatus(req.Status) {
		logging.LogWarn("admin update order failed: invalid order status", c, nil, "orderID", orderID, "status", req.Status)
		return HandleError(c, service.ErrInvalidOrderStatus)
	}

	if err := h.orderSvc.UpdateOrderStatus(uint(orderID), req.Status); err != nil {
		logging.LogWarn("admin update order failed: service error", c, err, "orderID", orderID)
		return HandleError(c, err)
	}

	logging.LogInfo("admin updated order successfully", c, "orderID", orderID, "status", req.Status)
	return c.JSON(fiber.Map{"message": "order status updated successfully"})
}
