package logger

import (
	"io"
	"log/slog"
	"os"
)

// SlogConfig holds all parameters needed to create a *slog.Logger.
type SlogConfig struct {
	// HandlerType specifies the built-in handler type: "json" or "text".
	// If empty, defaults to "json". Ignored if CustomHandler is non-nil.
	HandlerType string

	// Output destination for the handler. If nil, defaults to os.Stderr.
	Output io.Writer

	// Minimum logging level. If nil, defaults to slog.LevelInfo.
	Level slog.Leveler

	// HandlerOptions contains additional options like AddSource, ReplaceAttr.
	// If nil, a default with Level set is used.
	HandlerOptions *slog.HandlerOptions

	// CustomHandler, if non-nil, overrides all other fields and is used directly.
	CustomHandler slog.Handler
}

// ConfigOption is a functional option for modifying a SlogConfig.
type ConfigOption func(*SlogConfig)

// NewSlogConfig creates a new SlogConfig with defaults and applies the given options.
// Defaults: HandlerType="json", Output=os.Stderr, Level=Info, HandlerOptions=nil, CustomHandler=nil.
func NewSlogConfig(opts ...ConfigOption) SlogConfig {
	cfg := SlogConfig{
		HandlerType: "json",
		Output:      os.Stderr,
		Level:       slog.LevelInfo,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// SlogConfigDefault returns a SlogConfig that matches the default slog.Logger
// (text handler to stderr with level Info, no extra options).
func SlogConfigDefault() SlogConfig {
	return SlogConfig{
		HandlerType: "text",
		Output:      os.Stderr,
		Level:       slog.LevelInfo,
	}
}

// WithHandlerType returns a ConfigOption that sets the handler type.
func WithHandlerType(handlerType string) ConfigOption {
	return func(cfg *SlogConfig) {
		cfg.HandlerType = handlerType
	}
}

// WithOutput returns a ConfigOption that sets the output destination.
func WithOutput(w io.Writer) ConfigOption {
	return func(cfg *SlogConfig) {
		cfg.Output = w
	}
}

// WithLevel returns a ConfigOption that sets the minimum logging level.
func WithLevel(level slog.Level) ConfigOption {
	return func(cfg *SlogConfig) {
		cfg.Level = level
	}
}

// WithHandlerOptions returns a ConfigOption that sets HandlerOptions.
func WithHandlerOptions(opts *slog.HandlerOptions) ConfigOption {
	return func(cfg *SlogConfig) {
		cfg.HandlerOptions = opts
	}
}

// WithCustomHandler returns a ConfigOption that sets a custom handler,
// overriding all other settings.
func WithCustomHandler(handler slog.Handler) ConfigOption {
	return func(cfg *SlogConfig) {
		cfg.CustomHandler = handler
	}
}

// Clone returns a copy of the config.
func (c SlogConfig) Clone() SlogConfig {
	// Simple shallow copy; HandlerOptions and CustomHandler are shared,
	// but they are typically immutable after creation.
	return c
}

// newHandler creates a slog.Handler based on the configuration.
// It is unexported because it is only used internally.
func (c SlogConfig) newHandler() slog.Handler {
	if c.CustomHandler != nil {
		return c.CustomHandler
	}
	handlerType := c.HandlerType
	if handlerType == "" {
		handlerType = "json"
	}
	output := c.Output
	if output == nil {
		output = os.Stderr
	}
	level := c.Level
	if level == nil {
		level = slog.LevelInfo
	}
	opts := c.HandlerOptions
	if opts == nil {
		opts = &slog.HandlerOptions{Level: level}
	} else if opts.Level == nil {
		optsCopy := *opts
		optsCopy.Level = level
		opts = &optsCopy
	}
	switch handlerType {
	case "json":
		return slog.NewJSONHandler(output, opts)
	case "text":
		return slog.NewTextHandler(output, opts)
	default:
		return slog.NewJSONHandler(output, opts)
	}
}
