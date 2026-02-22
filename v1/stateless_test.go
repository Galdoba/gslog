package logger

import (
	"context"
	"log/slog"
	"testing"
)

func TestNewLogger(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	logger := cfg.NewLogger()
	logger.Info("hello")

	rec := th.lastRecord()
	if rec == nil {
		t.Fatal("no record logged")
	}
	if rec.Level != slog.LevelInfo {
		t.Errorf("level = %v, want %v", rec.Level, slog.LevelInfo)
	}
	if rec.Message != "hello" {
		t.Errorf("message = %q, want %q", rec.Message, "hello")
	}
}

func TestNewLoggerWithContext(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	fields := []ContextField{
		SimpleContextField("req_id"),
		ExtractorContextField("custom", func(ctx context.Context) any {
			return ctx.Value("custom_key")
		}),
	}
	logger := NewLoggerWithContext(cfg, fields, "")

	ctx := context.WithValue(context.Background(), "req_id", "123")
	ctx = context.WithValue(ctx, "custom_key", "value")
	logger.InfoContext(ctx, "msg")

	rec := th.lastRecord()
	if rec == nil {
		t.Fatal("no record")
	}
	attrs := flattenRecord(rec)
	if attrs["req_id"] != "123" {
		t.Errorf("req_id = %v, want %v", attrs["req_id"], "123")
	}
	if attrs["custom"] != "value" {
		t.Errorf("custom = %v, want %v", attrs["custom"], "value")
	}
}

func TestNewLoggerWithContextGroup(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	fields := []ContextField{SimpleContextField("req_id")}
	logger := NewLoggerWithContext(cfg, fields, "ctx")

	ctx := context.WithValue(context.Background(), "req_id", "123")
	logger.InfoContext(ctx, "msg")

	rec := th.lastRecord()
	if rec == nil {
		t.Fatal("no record")
	}
	attrs := flattenRecord(rec)
	if attrs["ctx.req_id"] != "123" {
		t.Errorf("ctx.req_id = %v, want %v", attrs["ctx.req_id"], "123")
	}
}
