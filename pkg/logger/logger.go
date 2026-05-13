package logger

import (
	"log/slog"
	"os"
)

// Field keys used across all log messages in the application.
const (
	KeyRequestID = "request_id"
	KeyMethod    = "method"
	KeyPath      = "path"
	KeyQuery     = "query"
	KeyStatus    = "status"
	KeyLatencyMs = "latency_ms"
	KeyIP        = "ip"
	KeyUserAgent = "user_agent"
	KeyBytesIn   = "bytes_in"
	KeyBytesOut  = "bytes_out"
	KeyError     = "error"
)

// Init configures the global slog logger with JSON output.
// Must be called before any logging occurs.
func Init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))
}
