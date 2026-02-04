package handler

import (
	"backend/handler/dto"
	"backend/service"
	"backend/utils/logging"

	"github.com/gofiber/fiber/v2"
)

type CartHandler struct {
	svc *service.CartService
}

func NewCartHandler(svc *service.CartService) *CartHandler {
	return &CartHandler{svc: svc}
}

func (h *CartHandler) Add(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var req dto.AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("add to cart failed: body parse", c, err, "userID", userID)
		return fiber.ErrBadRequest
	}

	if err := h.svc.AddItem(userID, req.ProductID, req.Quantity); err != nil {
		logging.LogWarn("add to cart failed: service error", c, err, "userID", userID, "productID", req.ProductID)
		return HandleError(c, err)
	}

	logging.LogInfo("add to cart succeeded", c, "userID", userID, "productID", req.ProductID, "quantity", req.Quantity)
	return c.SendStatus(fiber.StatusCreated)
}

func (h *CartHandler) Get(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	cart, err := h.svc.GetCart(userID)
	if err != nil {
		logging.LogWarn("get cart failed", c, err, "userID", userID)
		return HandleError(c, err)
	}

	logging.LogInfo("get cart succeeded", c, "userID", userID, "itemsCount", len(cart.Items))
	return c.JSON(cart)
}

func (h *CartHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	itemID, err := c.ParamsInt("itemId")
	if err != nil || itemID <= 0 {
		logging.LogWarn("update cart failed: invalid itemId", c, err, "userID", userID)
		return fiber.ErrBadRequest
	}

	var req dto.UpdateCartItems
	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("update cart failed: body parse", c, err, "userID", userID, "itemID", itemID)
		return fiber.ErrBadRequest
	}

	if err := h.svc.UpdateItem(userID, uint(itemID), req.Quantity); err != nil {
		logging.LogWarn("update cart failed: service error", c, err, "userID", userID, "itemID", itemID)
		return HandleError(c, err)
	}

	logging.LogInfo("update cart succeeded", c, "userID", userID, "itemID", itemID, "quantity", req.Quantity)
	return c.SendStatus(fiber.StatusOK)
}

func (h *CartHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	itemID, err := c.ParamsInt("itemId")
	if err != nil || itemID <= 0 {
		logging.LogWarn("delete cart item failed: invalid itemId", c, err, "userID", userID)
		return fiber.ErrBadRequest
	}

	if err := h.svc.RemoveItem(userID, uint(itemID)); err != nil {
		logging.LogWarn("delete cart item failed: service error", c, err, "userID", userID, "itemID", itemID)
		return HandleError(c, err)
	}

	logging.LogInfo("delete cart item succeeded", c, "userID", userID, "itemID", itemID)
	return c.SendStatus(fiber.StatusOK)
}
