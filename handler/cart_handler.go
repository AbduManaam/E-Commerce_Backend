package handler

import (
	"backend/handler/dto"
	"backend/service"

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
		return fiber.ErrBadRequest
	}

	if err := h.svc.AddItem(userID, req.ProductID, req.Quantity); err != nil {
		return HandleError(c, err)

	}

	return c.SendStatus(fiber.StatusCreated)
}

func (h *CartHandler) Get(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	cart, err := h.svc.GetCart(userID)
	if err != nil {
		return HandleError(c, err)
	}
	return c.JSON(cart)
}

func (h *CartHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	
	itemID, err := c.ParamsInt("itemId")
	if err != nil || itemID <= 0 {
		return fiber.ErrBadRequest
	}

	var req dto.UpdateCartItems
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	return h.svc.UpdateItem(userID, uint(itemID), req.Quantity)
}

func (h *CartHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	itemID, err := c.ParamsInt("itemId")
	if err != nil || itemID <= 0 {
		return fiber.ErrBadRequest
	}

	return h.svc.RemoveItem(userID, uint(itemID))
}