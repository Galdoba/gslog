package logger

import (
	"context"
	"log/slog"
)

// ContextExtractorHandler wraps a slog.Handler and adds fields extracted from the context
// to every log record. Fields are specified by a list of ContextField. If group is non-empty,
// all extracted fields are placed under that group.
type ContextExtractorHandler struct {
	next   slog.Handler
	fields []ContextField
	group  string
}

// Enabled reports whether the handler handles records at the given level.
// It delegates to the wrapped handler.
func (h *ContextExtractorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle handles the log record, adding context fields before passing to the wrapped handler.
func (h *ContextExtractorHandler) Handle(ctx context.Context, r slog.Record) error {
	var attrs []slog.Attr
	for _, f := range h.fields {
		var val any
		if f.Extractor != nil {
			val = f.Extractor(ctx)
		} else {
			val = ctx.Value(f.Key)
		}
		if val != nil {
			attrs = append(attrs, slog.Any(f.Key, val))
		}
	}
	if len(attrs) == 0 {
		return h.next.Handle(ctx, r)
	}
	if h.group != "" {
		anyAttrs := make([]any, len(attrs))
		for i, a := range attrs {
			anyAttrs[i] = a
		}
		r.AddAttrs(slog.Group(h.group, anyAttrs...))
	} else {
		for _, a := range attrs {
			r.AddAttrs(a)
		}
	}
	return h.next.Handle(ctx, r)
}

// WithAttrs returns a new handler whose attributes consist of the receiver's attributes
// combined with the given attributes, with context extraction preserved.
func (h *ContextExtractorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextExtractorHandler{
		next:   h.next.WithAttrs(attrs),
		fields: h.fields,
		group:  h.group,
	}
}

// WithGroup returns a new handler with the given group name, with context extraction preserved.
func (h *ContextExtractorHandler) WithGroup(name string) slog.Handler {
	return &ContextExtractorHandler{
		next:   h.next.WithGroup(name),
		fields: h.fields,
		group:  h.group,
	}
}

// WrapHandlerWithContext wraps an existing slog.Handler with a ContextExtractorHandler.
// The returned handler will add the specified fields from the context to every log record.
// If group is non-empty, the fields are grouped under that name.
func WrapHandlerWithContext(next slog.Handler, fields []ContextField, group string) slog.Handler {
	return &ContextExtractorHandler{
		next:   next,
		fields: fields,
		group:  group,
	}
}

// contextHandler is an unexported handler that adds a single value extracted from the context.
// It is used by Stateful.WithContextValue.
type contextHandler struct {
	next      slog.Handler
	key       string
	extractor func(context.Context) any
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if val := h.extractor(ctx); val != nil {
		r.AddAttrs(slog.Any(h.key, val))
	}
	return h.next.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{
		next:      h.next.WithAttrs(attrs),
		key:       h.key,
		extractor: h.extractor,
	}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{
		next:      h.next.WithGroup(name),
		key:       h.key,
		extractor: h.extractor,
	}
}
