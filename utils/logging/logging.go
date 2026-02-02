
package logging

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func Init(env string) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	Logger = slog.New(handler).
		With("app", "yumzy").
		With("env", env)
}
