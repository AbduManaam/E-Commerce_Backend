package handler

import (
	"backend/handler/dto"
	"backend/internal/domain"
	"backend/service"

	"github.com/gofiber/fiber/v2"
)

type AddressHandler struct {
	addressSvc *service.AddressService
}

func NewAddressHandler(addressSvc *service.AddressService) *AddressHandler {
	return &AddressHandler{addressSvc: addressSvc}
}

func (h *AddressHandler) Create(c *fiber.Ctx) error {
	uid, ok := c.Locals("userID").(uint)
	if !ok {
		return HandleError(c, service.ErrUnauthorized)
	}

	var req dto.CreateAddressRequest
	if err := c.BodyParser(&req); err != nil {
		return HandleError(c, service.ErrInvalidInput)
	}

	address := &domain.Address{
		FullName:  req.FullName,
		Phone:     req.Phone,
		Address:   req.Address,
		City:      req.City,
		State:     req.State,
		Country:   req.Country,
		ZipCode:   req.ZipCode,
		Landmark:  req.Landmark,
		IsDefault: req.IsDefault,
	}

	created, err := h.addressSvc.Create(uid, address)
	if err != nil {
		return HandleError(c, err)
	}

	if created.IsDefault {
		if err := h.addressSvc.UnsetOtherDefaults(uid, created.ID); err != nil {
			return HandleError(c, err)
		}
	}

	return c.Status(201).JSON(created)
}
