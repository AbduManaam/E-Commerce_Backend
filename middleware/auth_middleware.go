package middleware

import (
	"backend/config"
	"backend/repository"
	"backend/utils"
	"backend/utils/logging"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(cfg *config.AppConfig, userRepo repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID, _ := c.Locals(logging.RequestIDKey).(string)

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logging.LogWarn("missing authorization header",
				"request_id", requestID,
				"path", c.Path(),
				"ip", c.IP(),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		// Validate format: "Bearer <token>"
		if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			logging.LogWarn("invalid authorization format",
				"request_id", requestID,
				"path", c.Path(),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization format",
			})
		}

		tokenStr := authHeader[7:]

		// Validate the access token
		claims, err := utils.ValidateAccessToken(tokenStr, cfg.JWT.AccessSecret)
		if err != nil {
			logging.LogWarn("token validation failed",
				"request_id", requestID,
				"error", err.Error(),
				"path", c.Path(),
			)

			// Handle specific JWT errors
			if errors.Is(err, jwt.ErrTokenExpired) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "access token expired",
				})
			}
			if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "invalid token signature",
				})
			}

			// Fallback for other errors
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		// Load user from DB
		user, err := userRepo.GetByID(claims.UserID)
		if err != nil || user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user not found",
			})
		}

		// Check if user is blocked
		if user.IsBlocked {
			c.ClearCookie("access_token")
			c.ClearCookie("refresh_token")

			logging.LogWarn("blocked user access attempt",
				"request_id", requestID,
				"user_id", user.ID,
				"path", c.Path(),
			)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Account suspended. Contact support.",
				"code":  "USER_BLOCKED",
			})
		}

		// Attach user info to context
		c.Locals(logging.UserIDKey, claims.UserID)
		c.Locals(logging.RoleKey, claims.Role)
		c.Locals("isAdmin", claims.Role == "admin")

		return c.Next()
	}
}
