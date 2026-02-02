package logging

import "github.com/gofiber/fiber/v2"

func LogWarn(msg string, c *fiber.Ctx, err error, extra ...any) {
	fields := []any{
		"error", err.Error(),
	}

	if c != nil {
		fields = append(fields,
			"path", c.Path(),
			"method", c.Method(),
			"ip", c.IP(),
		)
	}

	fields = append(fields, extra...)
	Logger.Warn(msg, fields...)
}

func LogInfo(msg string, c *fiber.Ctx, extra ...any) {
	fields := []any{}

	if c != nil {
		fields = append(fields,
			"path", c.Path(),
			"method", c.Method(),
			"ip", c.IP(),
		)
	}

	fields = append(fields, extra...)
	Logger.Info(msg, fields...)
}
