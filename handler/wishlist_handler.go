package handler

import (
	"backend/handler/dto"
	"backend/service"
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
        return HandleError(c, service.ErrInvalidInput)
    }

    if err := validator.Validate.Struct(req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "errors": validator.FormatErrors(err),
        })
    }

    if err := h.wishlistSvc.Add(userID, req.ProductID); err != nil {
        return HandleError(c, err)
    }

    return c.JSON(fiber.Map{"message": "product added to wishlist"})
}


// func (h *WishlistHandler) Add(c *fiber.Ctx) error {
// 	var req struct {
// 		ProductID uint `json:"productID"`
// 	}

// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON"})
// 	}

// 	if req.ProductID == 0 {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "productID is required"})
// 	}

// 	if err := h.wishlistSvc.Add( /* get userID from context */ c.Locals("userID").(uint), req.ProductID); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.JSON(fiber.Map{"message": "product added to wishlist"})
// }

// Get wishlist
func (h *WishlistHandler) Get(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	items, err := h.wishlistSvc.Get(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

// Remove product from wishlist
func (h *WishlistHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	productID, err := c.ParamsInt("product_id")
	if err != nil || productID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid product_id"})
	}

	if err := h.wishlistSvc.Remove(userID, uint(productID)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "product removed from wishlist"})
}
