package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestContextExtractorHandler(t *testing.T) {
	th := newTestHandler()
	fields := []ContextField{
		SimpleContextField("req_id"),
		ExtractorContextField("custom", func(ctx context.Context) any {
			return ctx.Value("custom_key")
		}),
	}
	handler := WrapHandlerWithContext(th, fields, "")
	logger := slog.New(handler)

	ctx := context.WithValue(context.Background(), "req_id", "123")
	ctx = context.WithValue(ctx, "custom_key", "value")
	logger.InfoContext(ctx, "msg")

	rec := th.lastRecord()
	if rec == nil {
		t.Fatal("no record")
	}
	attrs := flattenRecord(rec)
	if attrs["req_id"] != "123" {
		t.Errorf("req_id = %v, want 123", attrs["req_id"])
	}
	if attrs["custom"] != "value" {
		t.Errorf("custom = %v, want value", attrs["custom"])
	}
}

func TestContextExtractorHandlerWithGroup(t *testing.T) {
	th := newTestHandler()
	fields := []ContextField{SimpleContextField("req_id")}
	handler := WrapHandlerWithContext(th, fields, "ctx")
	logger := slog.New(handler)

	ctx := context.WithValue(context.Background(), "req_id", "123")
	logger.InfoContext(ctx, "msg")

	rec := th.lastRecord()
	attrs := flattenRecord(rec)
	if attrs["ctx.req_id"] != "123" {
		t.Errorf("ctx.req_id = %v, want 123", attrs["ctx.req_id"])
	}
}

func TestContextExtractorHandlerEnabled(t *testing.T) {
	th := newTestHandler()
	handler := WrapHandlerWithContext(th, nil, "")
	if !handler.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("Enabled returned false")
	}
}

func TestContextExtractorHandlerWithAttrs(t *testing.T) {
	th := newTestHandler() // was &testHandler{}
	fields := []ContextField{SimpleContextField("req_id")}
	handler := WrapHandlerWithContext(th, fields, "")
	handler2 := handler.WithAttrs([]slog.Attr{slog.String("static", "value")})
	_, ok := handler2.(*ContextExtractorHandler)
	if !ok {
		t.Fatal("WithAttrs did not return ContextExtractorHandler")
	}
	logger := slog.New(handler2)
	ctx := context.WithValue(context.Background(), "req_id", "123")
	logger.InfoContext(ctx, "msg")

	rec := th.lastRecord()
	attrs := flattenRecord(rec)
	if attrs["req_id"] != "123" {
		t.Error("req_id missing")
	}
	if attrs["static"] != "value" {
		t.Error("static missing")
	}
}

func TestContextExtractorHandlerWithGroupChaining(t *testing.T) {
	var buf bytes.Buffer
	baseHandler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})

	fields := []ContextField{SimpleContextField("req_id")}
	handler := WrapHandlerWithContext(baseHandler, fields, "ctx")
	handler2 := handler.WithGroup("app")
	logger := slog.New(handler2)

	ctx := context.WithValue(context.Background(), "req_id", "123")
	logger.InfoContext(ctx, "msg")

	output := buf.String()
	// Now parse or search for the expected key-value pair.
	// For JSON handler, you could unmarshal; for text, you can use string contains.
	if !strings.Contains(output, "app.ctx.req_id=123") {
		t.Errorf("expected app.ctx.req_id=123 in output, got: %s", output)
	}
}
