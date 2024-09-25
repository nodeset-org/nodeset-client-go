package common

import "log/slog"

// Logs a message at the debug level if the logger is not nil
func SafeDebugLog(logger *slog.Logger, msg string, args ...any) {
	if logger == nil {
		return
	}
	logger.Debug(msg, args...)
}
