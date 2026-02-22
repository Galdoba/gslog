package logger

import (
	"log/slog"
)

// NewLogger creates a simple *slog.Logger with the configuration applied.
func (c SlogConfig) NewLogger() *slog.Logger {
	return slog.New(c.newHandler())
}

// Stateless creates a plain slog.Logger from the configuration.
// It is an alias for NewLogger for clarity.
func (c SlogConfig) Stateless() *slog.Logger {
	return c.NewLogger()
}

// NewLoggerWithContext creates a *slog.Logger based on the configuration,
// wrapped with a ContextExtractorHandler. The returned logger will add the
// specified fields from the context to every log record. If group is non-empty,
// the fields are grouped under that name.
func NewLoggerWithContext(cfg SlogConfig, fields []ContextField, group string) *slog.Logger {
	baseHandler := cfg.newHandler()
	handler := WrapHandlerWithContext(baseHandler, fields, group)
	return slog.New(handler)
}
