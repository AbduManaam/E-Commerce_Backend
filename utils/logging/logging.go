package logging

import (
	"log/slog"
	"os"
)

// Logger is the global structured logger instance.
var Logger *slog.Logger

// Init initializes the global logger based on the environment.
//   - "development" → human-readable text output at DEBUG level
//   - anything else  → JSON output at INFO level (production-ready for ELK/Grafana/Loki/Datadog)
//
// An optional log level override can be provided (e.g., "debug", "info", "warn", "error").
func Init(env string, levelOverride ...string) {
	var level slog.Level

	// Determine base level from environment
	if env == "development" {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	// Apply optional level override
	if len(levelOverride) > 0 && levelOverride[0] != "" {
		switch levelOverride[0] {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		}
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if env == "development" {
		// Human-readable text format for terminal debugging
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		// Structured JSON for production log aggregation
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	Logger = slog.New(handler).With("service", "yumzy-api")
}

// GetLogger returns the global logger instance.
// If Init has not been called, it initializes a default logger.
func GetLogger() *slog.Logger {
	if Logger == nil {
		Init("development")
	}
	return Logger
}
