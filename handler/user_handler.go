package handler

import (
	"backend/handler/dto"
	"backend/service"
	validator "backend/utils/validation"
	"backend/utils/logging"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	user, err := h.userSvc.GetProfile(userID)
	if err != nil {
		logging.LogWarn("get profile failed: service error", c, err, "userID", userID)
		return HandleError(c, err)
	}

	logging.LogInfo("user profile retrieved successfully", c, "userID", userID)

	resp := dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	return c.JSON(resp)
}

func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var req struct {
		Name string `json:"name" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("update profile failed: body parse", c, err, "userID", userID)
		return HandleError(c, service.ErrInvalidInput)
	}

	if err := validator.Validate.Struct(req); err != nil {
		logging.LogWarn("update profile failed: validation error", c, err, "userID", userID)
		return c.Status(400).JSON(fiber.Map{
			"errors": validator.FormatErrors(err),
		})
	}

	if err := h.userSvc.UpdateProfile(userID, req.Name); err != nil {
		logging.LogWarn("update profile failed: service error", c, err, "userID", userID)
		return HandleError(c, err)
	}

	logging.LogInfo("user profile updated successfully", c, "userID", userID)
	return c.JSON(fiber.Map{"message": "Profile updated successfully"})
}
