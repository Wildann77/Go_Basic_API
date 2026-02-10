package logger

import (
	"context"
	"log/slog"
	"os"
)

var Log *slog.Logger

func Init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize attributes if needed, e.g., rename time key or format
			return a
		},
	}

	// Use JSON Handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, opts)
	Log = slog.New(handler)

	// Set as default logger
	slog.SetDefault(Log)
}

// WithContext returns a logger with context attributes (like RequestID)
func WithContext(ctx context.Context) *slog.Logger {
	// Example: Extract RequestID from context if available
	// Assuming middleware sets "RequestID" in context
	reqID, ok := ctx.Value("RequestID").(string)
	if ok {
		return Log.With(slog.String("request_id", reqID))
	}
	return Log
}

// Info logs at Info level using standard logger
func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

// Error logs at Error level using standard logger
func Error(msg string, args ...any) {
	Log.Error(msg, args...)
}

// Debug logs at Debug level using standard logger
func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

// Warn logs at Warn level using standard logger
func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}

// Fatal logs at Error level and exits
func Fatal(msg string, args ...any) {
	Log.Error(msg, args...)
	os.Exit(1)
}
