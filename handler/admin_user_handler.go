package handler

import (
	"strconv"

	"backend/handler/dto"
	"backend/service"
	"backend/utils/logging"
	validator "backend/utils/validation"

	"github.com/gofiber/fiber/v2"
)


type AdminUserHandler struct {
	userSvc *service.UserService
	orderSvc *service.OrderService
}

func NewAdminUserHandler(userSvc *service.UserService,orderSvc *service.OrderService,) *AdminUserHandler {
	return &AdminUserHandler{
		userSvc: userSvc,
		orderSvc: orderSvc}
}

//BlockUser
func (h *AdminUserHandler) BlockUser(c *fiber.Ctx) error {
	idParam := c.Params("id")
	userID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		logging.LogWarn("block user failed: invalid id", c, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	isAdmin, ok := c.Locals("isAdmin").(bool)
if !ok || !isAdmin {
    logging.LogWarn("block user failed: admin privileges required", c,err, "userID", userID)
    return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
        "error": "Only admins can block users",
    })
}

if err := h.userSvc.BlockUser(isAdmin, uint(userID)); err != nil {
    logging.LogWarn("block user failed: service error", c, err, "userID", userID)
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
        "error": "Unable to block user due to internal error",
    })
}

logging.LogInfo("user blocked successfully", c, "userID", userID)
return c.JSON(fiber.Map{"message": "user blocked successfully"})

}

func (h *AdminUserHandler) UpdateUser(c *fiber.Ctx) error {
	idParam := c.Params("id")
	userID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		logging.LogWarn("update user failed: invalid ID", c, err)
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}

	var req dto.AdminUpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("update user failed: body parse", c, err, "userID", userID)
		return c.Status(400).JSON(fiber.Map{  "error": "Invalid request body. Please check the input format."})
	}

	if err:= validator.Validate.Struct(req);err!=nil{
		return c.Status(400).JSON(fiber.Map{
			"errors":validator.FormatErrors(err),
		})
	}

	isAdmin := c.Locals("isAdmin").(bool)
	if err := h.userSvc.AdminUpdateUser(isAdmin, uint(userID), req.Name, req.Role); err != nil {
		logging.LogWarn("update user failed: service error", c, err, "userID", userID, "role", req.Role)
		return c.Status(403).JSON(fiber.Map{ "error": "Only admins can block users"})
	}

	logging.LogInfo("user updated successfully", c, "userID", userID, "name", req.Name, "role", req.Role)
	return c.JSON(fiber.Map{"message": "user updated successfully"})
}
func (h *AdminUserHandler ) ListUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	if page < 1 || limit < 1 || limit > 100 {
		return HandleError(c, service.ErrInvalidInput)
	}

	users, total, err := h.userSvc.ListAllWithCount(page, limit)
	if err != nil {
		return HandleError(c, err)
	}

	return c.JSON(fiber.Map{
		"data":  users,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

func (h *AdminUserHandler ) GetUserOrders(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("user_id")
	if err != nil || userID <= 0 {
		return HandleError(c, service.ErrInvalidInput)
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	if page < 1 || limit < 1 || limit > 100 {
		return HandleError(c, service.ErrInvalidInput)
	}

	orders, total, err := h.orderSvc.GetUserOrdersPaginated(
		uint(userID),
		page,
		limit,
	)
	if err != nil {
		return HandleError(c, err)
	}

	return c.JSON(fiber.Map{
		"data":  orders,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

func (h *AdminUserHandler ) GetOrderDetails(c *fiber.Ctx) error {
	orderID, err := c.ParamsInt("order_id")
	if err != nil || orderID <= 0 {
		return HandleError(c, service.ErrInvalidInput)
	}

	order, err := h.orderSvc.GetOrderDetail(uint(orderID))
	if err != nil {
		return HandleError(c, err)
	}

	return c.JSON(order)
}