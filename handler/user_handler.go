package handler

import (
	"backend/handler/dto"
	"backend/service"
	validator "backend/utils/validation"

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
		return HandleError(c, err)
	}

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
		return HandleError(c, service.ErrInvalidInput)
	}
	if err:= validator.Validate.Struct(req);err!=nil{
		return c.Status(400).JSON(fiber.Map{
			"errors":validator.FormatErrors(err),
		})
	}


	if err := h.userSvc.UpdateProfile(userID, req.Name); err != nil {
		return HandleError(c, err)
	}

	return c.JSON(fiber.Map{"message": "Profile updated successfully"})
}