package logger

import (
	"io"
	"log/slog"
	"os"
	"testing"
)

func TestNewSlogConfigDefaults(t *testing.T) {
	cfg := NewSlogConfig()
	if cfg.HandlerType != "json" {
		t.Errorf("HandlerType = %q, want %q", cfg.HandlerType, "json")
	}
	if cfg.Output != os.Stderr {
		t.Errorf("Output = %v, want %v", cfg.Output, os.Stderr)
	}
	if cfg.Level != slog.LevelInfo {
		t.Errorf("Level = %v, want %v", cfg.Level, slog.LevelInfo)
	}
	if cfg.HandlerOptions != nil {
		t.Errorf("HandlerOptions = %v, want nil", cfg.HandlerOptions)
	}
	if cfg.CustomHandler != nil {
		t.Errorf("CustomHandler = %v, want nil", cfg.CustomHandler)
	}
}

func TestSlogConfigDefault(t *testing.T) {
	cfg := SlogConfigDefault()
	if cfg.HandlerType != "text" {
		t.Errorf("HandlerType = %q, want %q", cfg.HandlerType, "text")
	}
	if cfg.Output != os.Stderr {
		t.Errorf("Output = %v, want %v", cfg.Output, os.Stderr)
	}
	if cfg.Level != slog.LevelInfo {
		t.Errorf("Level = %v, want %v", cfg.Level, slog.LevelInfo)
	}
}

func TestWithHandlerType(t *testing.T) {
	cfg := NewSlogConfig(WithHandlerType("text"))
	if cfg.HandlerType != "text" {
		t.Errorf("HandlerType = %q, want %q", cfg.HandlerType, "text")
	}
}

func TestWithOutput(t *testing.T) {
	var buf io.Writer // nil is acceptable for testing
	cfg := NewSlogConfig(WithOutput(buf))
	if cfg.Output != buf {
		t.Errorf("Output = %v, want %v", cfg.Output, buf)
	}
}

func TestWithLevel(t *testing.T) {
	cfg := NewSlogConfig(WithLevel(slog.LevelDebug))
	if cfg.Level != slog.LevelDebug {
		t.Errorf("Level = %v, want %v", cfg.Level, slog.LevelDebug)
	}
}

func TestWithHandlerOptions(t *testing.T) {
	opts := &slog.HandlerOptions{AddSource: true}
	cfg := NewSlogConfig(WithHandlerOptions(opts))
	if cfg.HandlerOptions != opts {
		t.Errorf("HandlerOptions = %v, want %v", cfg.HandlerOptions, opts)
	}
}

func TestWithCustomHandler(t *testing.T) {
	handler := slog.NewTextHandler(os.Stderr, nil)
	cfg := NewSlogConfig(WithCustomHandler(handler))
	if cfg.CustomHandler != handler {
		t.Errorf("CustomHandler = %v, want %v", cfg.CustomHandler, handler)
	}
}

func TestClone(t *testing.T) {
	orig := NewSlogConfig(WithHandlerType("text"), WithLevel(slog.LevelWarn))
	clone := orig.Clone()
	if clone.HandlerType != orig.HandlerType {
		t.Errorf("Clone HandlerType = %q, want %q", clone.HandlerType, orig.HandlerType)
	}
	if clone.Level != orig.Level {
		t.Errorf("Clone Level = %v, want %v", clone.Level, orig.Level)
	}
	// Modify original; clone should not be affected (shallow copy)
	orig.HandlerType = "json"
	if clone.HandlerType == "json" {
		t.Error("Clone was affected by modification of original")
	}
}
