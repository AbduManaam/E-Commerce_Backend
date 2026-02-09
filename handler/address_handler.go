package handler

import (
	"strconv"

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
	uid := c.Locals("userID").(uint)

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

	return c.Status(201).JSON(created)
}

func (h *AddressHandler) List(c *fiber.Ctx) error {
	uid := c.Locals("userID").(uint)

	addresses, err := h.addressSvc.List(uid)
	if err != nil {
		return HandleError(c, err)
	}

	return c.JSON(addresses)
}

func (h *AddressHandler) GetByID(c *fiber.Ctx) error {
	uid := c.Locals("userID").(uint)

	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	address, err := h.addressSvc.GetByID(uid, uint(id))
	if err != nil {
		return HandleError(c, err)
	}

	return c.JSON(address)
}

func (h *AddressHandler) Update(c *fiber.Ctx) error {
	uid := c.Locals("userID").(uint)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var req dto.CreateAddressRequest
	if err := c.BodyParser(&req); err != nil {
		return HandleError(c, service.ErrInvalidInput)
	}

	address := &domain.Address{
		ID:        uint(id),
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

	if err := h.addressSvc.Update(uid, address); err != nil {
		return HandleError(c, err)
	}

	return c.JSON(address)
}

func (h *AddressHandler) Delete(c *fiber.Ctx) error {
	uid := c.Locals("userID").(uint)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	if err := h.addressSvc.Delete(uid, uint(id)); err != nil {
		return HandleError(c, err)
	}

	return c.JSON(fiber.Map{"message": "address deleted"})
}

func (h *AddressHandler) SetDefault(c *fiber.Ctx) error {
	uid := c.Locals("userID").(uint)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	if err := h.addressSvc.SetDefault(uid, uint(id)); err != nil {
		return HandleError(c, err)
	}

	return c.JSON(fiber.Map{"message": "default address updated"})
}
