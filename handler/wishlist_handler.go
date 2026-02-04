package handler

import (
	"backend/handler/dto"
	"backend/service"
	"backend/utils/logging"
	validator "backend/utils/validation"

	"github.com/gofiber/fiber/v2"
)

type WishlistHandler struct {
	wishlistSvc *service.WishlistService
}

func NewWishlistHandler(svc *service.WishlistService) *WishlistHandler {
	return &WishlistHandler{wishlistSvc: svc}
}

// Add product to wishlist
func (h *WishlistHandler) Add(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var req dto.WishlistRequest
	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("add wishlist failed: body parse", c, err, "userID", userID)
		return HandleError(c, service.ErrInvalidInput)
	}

	if err := validator.Validate.Struct(req); err != nil {
		logging.LogWarn("add wishlist failed: validation error", c, err, "userID", userID, "productID", req.ProductID)
		return c.Status(400).JSON(fiber.Map{
			"errors": validator.FormatErrors(err),
		})
	}

	if err := h.wishlistSvc.Add(userID, req.ProductID); err != nil {
		logging.LogWarn("add wishlist failed: service error", c, err, "userID", userID, "productID", req.ProductID)
		return HandleError(c, err)
	}

	logging.LogInfo("product added to wishlist successfully", c, "userID", userID, "productID", req.ProductID)
	return c.JSON(fiber.Map{"message": "product added to wishlist"})
}

// Get wishlist
func (h *WishlistHandler) Get(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	items, err := h.wishlistSvc.Get(userID)
	if err != nil {
		logging.LogWarn("get wishlist failed: service error", c, err, "userID", userID)
		return HandleError(c, err)
	}

	logging.LogInfo("wishlist retrieved successfully", c, "userID", userID, "itemsCount", len(items))
	return c.JSON(items)
}

// Remove product from wishlist
func (h *WishlistHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	productID, err := c.ParamsInt("product_id")
	if err != nil || productID <= 0 {
		logging.LogWarn("delete wishlist failed: invalid product_id", c, err, "userID", userID, "productIDParam", c.Params("product_id"))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid product_id"})
	}

	if err := h.wishlistSvc.Remove(userID, uint(productID)); err != nil {
		logging.LogWarn("delete wishlist failed: service error", c, err, "userID", userID, "productID", productID)
		return HandleError(c, err)
	}

	logging.LogInfo("product removed from wishlist successfully", c, "userID", userID, "productID", productID)
	return c.JSON(fiber.Map{"message": "product removed from wishlist"})
}
