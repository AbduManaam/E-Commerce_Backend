package logging

import "os"

// LogInfo logs an informational message with optional structured key-value fields.
func LogInfo(msg string, fields ...any) {
	GetLogger().Info(msg, fields...)
}

// LogDebug logs a debug-level message with optional structured key-value fields.
func LogDebug(msg string, fields ...any) {
	GetLogger().Debug(msg, fields...)
}

// LogWarn logs a warning message with optional structured key-value fields.
func LogWarn(msg string, fields ...any) {
	GetLogger().Warn(msg, fields...)
}

// LogError logs an error-level message with optional structured key-value fields.
func LogError(msg string, fields ...any) {
	GetLogger().Error(msg, fields...)
}

// LogFatal logs an error-level message and exits the process.
// Use for unrecoverable startup errors only.
func LogFatal(msg string, fields ...any) {
	GetLogger().Error(msg, fields...)
	os.Exit(1)
}
