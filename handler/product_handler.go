package handler

import (
	"strconv"

	"backend/handler/dto"
	"backend/service"
	"backend/utils/logging"
	validator "backend/utils/validation"

	"github.com/gofiber/fiber/v2"
)

type ProductHandler struct {
	productSvc *service.ProductService
}

func NewProductHandler(productSvc *service.ProductService) *ProductHandler {
	return &ProductHandler{productSvc: productSvc}
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req dto.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("create product failed: body parse", c, err)
		return HandleError(c, service.ErrInvalidInput)
	}
	if err:= validator.Validate.Struct(req);err!=nil{
		return c.Status(400).JSON(fiber.Map{
			"errors":validator.FormatErrors(err),
		})
	}


	product, err := h.productSvc.CreateProduct(req)
	if err != nil {
		logging.LogWarn("create product failed: service error", c, err, "name", req.Name)
		return HandleError(c, err)
	}

	logging.LogInfo("product created successfully", c, "productID", product.ID)
	return c.Status(201).JSON(product)
}


func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	idParam := c.Params("id")
	productID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return HandleError(c, service.ErrInvalidInput)
	}

	product, err := h.productSvc.GetProduct(uint(productID))
	if err != nil {
		return HandleError(c, err)
	}

	return c.JSON(product)
}


func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	idParam := c.Params("id")
	productID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		logging.LogWarn("update product failed: invalid product ID", c, err)
		return HandleError(c, service.ErrInvalidInput)
	}

		var req dto.UpdateProductRequest

	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("update product failed: body parse", c, err, "productID", productID)
		return HandleError(c, service.ErrInvalidInput)
	}

	if err:= validator.Validate.Struct(req);err!=nil{
		return c.Status(400).JSON(fiber.Map{
			"errors":validator.FormatErrors(err),
		})
	}


	if err := h.productSvc.UpdateProduct(uint(productID), req); err != nil {
		return HandleError(c, err)
	}

	logging.LogInfo("product updated successfully", c, "productID", productID)
	return c.JSON(fiber.Map{"message": "Product updated successfully"})
}



func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	idParam := c.Params("id")
	productID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		logging.LogWarn("delete product failed: invalid product ID", c, err)
		return HandleError(c, service.ErrInvalidInput)
	}

	if err := h.productSvc.DeleteProduct(uint(productID)); err != nil {
		logging.LogWarn("delete product failed: service error", c, err, "productID", productID)
		return HandleError(c, err)
	}

	logging.LogInfo("product deleted successfully", c, "productID", productID)
	return c.JSON(fiber.Map{"message": "Product deleted successfully"})
}

func (h *ProductHandler) ListProducts(c *fiber.Ctx) error {
	products, err := h.productSvc.ListProducts()
	if err != nil {
		logging.LogWarn("list products failed", c, err)
		return HandleError(c, err)
	}

	logging.LogInfo("products listed successfully", c, "count", len(products))
	return c.JSON(products)
}
