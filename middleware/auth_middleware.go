package middleware

import (
	"strings"

	"backend/config"
	"backend/repository"
	"backend/utils"
	"backend/utils/logging"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(cfg *config.AppConfig,userRepo repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logging.LogWarn("missing authorization header", c, fiber.ErrUnauthorized)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logging.LogWarn("invalid authorization format", c, fiber.ErrUnauthorized)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization format",
			})
		}

		claims, err := utils.ValidateAccessToken(parts[1], cfg.JWT.AccessSecret)
		if err != nil {
			logging.LogWarn("invalid or expired access token", c, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}
			// 3. Load user from DB (CRITICAL)
		user, err := userRepo.GetByID(claims.UserID)
		if err != nil || user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user not found",
			})
		}

		// 4. Blocked user check (FORCE LOGOUT)
		if user.IsBlocked {
			c.ClearCookie("access_token")
			c.ClearCookie("refresh_token")

			logging.LogWarn("blocked user access attempt", c, nil,
				"userID", user.ID,
				"ip", c.IP(),
			)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Account suspended. Contact support.",
				"code":  "USER_BLOCKED",
			})
		}

		// Attaching user info to context
		c.Locals("userID", claims.UserID)
		c.Locals("role", claims.Role)
		c.Locals("isAdmin", claims.Role == "admin")

		logging.LogInfo("authenticated request", c,
			"userID", claims.UserID,
			"role", claims.Role,
		)

		return c.Next()
	}
}
