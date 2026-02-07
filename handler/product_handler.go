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

	if err := validator.Validate.Struct(req); err != nil {
		logging.LogWarn("create product failed: validation error", c, err, "name", req.Name)
		return c.Status(400).JSON(fiber.Map{
			"errors": validator.FormatErrors(err),
		})
	}

	product, err := h.productSvc.CreateProduct(req)
	if err != nil {
		logging.LogWarn("create product failed: service error", c, err, "name", req.Name)
		return HandleError(c, err)
	}

	logging.LogInfo("product created successfully", c, "productID", product.ID, "name", req.Name)
	return c.Status(201).JSON(product)
}

func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	idParam := c.Params("id")
	productID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		logging.LogWarn("get product failed: invalid product ID", c, err, "productIDParam", idParam)
		return HandleError(c, service.ErrInvalidInput)
	}

	product, err := h.productSvc.GetProduct(uint(productID))
	if err != nil {
		logging.LogWarn("get product failed: service error", c, err, "productID", productID)
		return HandleError(c, err)
	}

	logging.LogInfo("product retrieved successfully", c, "productID", productID)
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

	if err := validator.Validate.Struct(req); err != nil {
		logging.LogWarn("update product failed: validation error", c, err, "productID", productID)
		return c.Status(400).JSON(fiber.Map{
			"errors": validator.FormatErrors(err),
		})
	}

	if err := h.productSvc.UpdateProduct(uint(productID), req); err != nil {
		logging.LogWarn("update product failed: service error", c, err, "productID", productID)
		return HandleError(c, err)
	}

	logging.LogInfo("product updated successfully", c, "productID", productID)
	return c.JSON(fiber.Map{"message": "Product updated successfully"})
}

func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	idParam := c.Params("id")
	productID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		logging.LogWarn("delete product failed: invalid product ID", c, err, "productIDParam", idParam)
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
		logging.LogWarn("list products failed: service error", c, err)
		return HandleError(c, err)
	}

	logging.LogInfo("products listed successfully", c, "count", len(products))
	return c.JSON(products)
}

func (h *ProductHandler) ListFiltered(c *fiber.Ctx) error {

	var minPrice *float64

	if c.Query("min_price") !=""{
      v:= c.QueryFloat("min_price")
	  minPrice= &v
	}
	var maxPrice *float64

	if c.Query("max_price") !=""{
      v:= c.QueryFloat("max_price")
	  maxPrice= &v
	}

	var categoryId *uint


	if c.Query("category_id")!=""{
      v:= c.QueryInt("category_id")
	  if  v>0{
		val:=uint(v)
		categoryId= &val
	  }
	}

	req := dto.ProductListQuery{
		CategoryID: categoryId,
		Sort:     c.Query("sort","created_at"),
		Order:    c.Query("order", "desc"),
		Search:   c.Query("search"),
		Page:     c.QueryInt("page", 1),
		Limit:    c.QueryInt("limit", 10),
		MinPrice: minPrice,
		MaxPrice: maxPrice,
	}

	products, err := h.productSvc.ListActive(req)
	if err != nil {
		return HandleError(c, err)
	}

	return c.JSON(products)
}

//-----------------------------------------

type CategoryHandler struct {
	service *service.CategoryService
}

func NewCategoryHandler(service *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	var body struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid JSON body"})
	}

	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "category name is required"})
	}

	category, err := h.service.Create(body.Name)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(category)
}

func (h *CategoryHandler) List(c *fiber.Ctx) error {
	categories, err := h.service.List()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(categories)
}
