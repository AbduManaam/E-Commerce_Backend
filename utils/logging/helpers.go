package logging

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
