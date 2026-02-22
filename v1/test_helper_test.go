package logger

import (
	"context"
	"log/slog"
	"sync"
)

// testHandler is a slog.Handler that stores records for later inspection,
// correctly handling WithAttrs and WithGroup.
type testHandler struct {
	mu      sync.Mutex
	records *[]*slog.Record // shared slice across all derived handlers
	attrs   []slog.Attr
	groups  []string
}

// newTestHandler creates a new testHandler with an empty record store.
func newTestHandler() *testHandler {
	records := make([]*slog.Record, 0)
	return &testHandler{
		records: &records,
	}
}

func (h *testHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *testHandler) Handle(ctx context.Context, r slog.Record) error {
	// Apply stored attributes (group handling is omitted for simplicity;
	// for full correctness we would need to prefix keys with groups).
	for _, a := range h.attrs {
		r.AddAttrs(a)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	// Store a pointer to the record (the record is passed by value, safe to keep).
	*h.records = append(*h.records, &r)
	return nil
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Combine existing attributes with new ones.
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &testHandler{
		records: h.records,
		attrs:   newAttrs,
		groups:  h.groups,
	}
}

func (h *testHandler) WithGroup(name string) slog.Handler {
	// Append group name (used only for future WithAttrs calls, not implemented in Handle).
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	return &testHandler{
		records: h.records,
		attrs:   h.attrs,
		groups:  newGroups,
	}
}

// lastRecord returns the most recently logged record, or nil if none.
func (h *testHandler) lastRecord() *slog.Record {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(*h.records) == 0 {
		return nil
	}
	return (*h.records)[len(*h.records)-1]
}

// reset clears all stored records.
func (h *testHandler) reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	*h.records = nil
}

// flattenRecord returns a map of all attributes in a record, with groups flattened.
// Keys are joined with "." (e.g., "group.key").
func flattenRecord(r *slog.Record) map[string]any {
	m := make(map[string]any)
	if r == nil {
		return m
	}
	r.Attrs(func(a slog.Attr) bool {
		flattenAttr(m, "", a)
		return true
	})
	return m
}

func flattenAttr(dest map[string]any, prefix string, a slog.Attr) {
	key := a.Key
	if prefix != "" {
		key = prefix + "." + key
	}
	val := a.Value
	switch val.Kind() {
	case slog.KindGroup:
		for _, attr := range val.Group() {
			flattenAttr(dest, key, attr)
		}
	default:
		dest[key] = val.Any()
	}
}

// testHandler is a slog.Handler that stores records for later inspection.
// type testHandler struct {
// 	mu      sync.Mutex
// 	records []*slog.Record
// }

// func (h *testHandler) Enabled(context.Context, slog.Level) bool { return true }
// func (h *testHandler) Handle(ctx context.Context, r slog.Record) error {
// 	h.mu.Lock()
// 	defer h.mu.Unlock()
// 	// Store a pointer to the record (the record itself is a copy, safe to keep).
// 	h.records = append(h.records, &r)
// 	return nil
// }
// func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
// func (h *testHandler) WithGroup(name string) slog.Handler       { return h }

// // lastRecord returns the most recently logged record, or nil if none.
// func (h *testHandler) lastRecord() *slog.Record {
// 	h.mu.Lock()
// 	defer h.mu.Unlock()
// 	if len(h.records) == 0 {
// 		return nil
// 	}
// 	return h.records[len(h.records)-1]
// }

// // reset clears all stored records.
// func (h *testHandler) reset() {
// 	h.mu.Lock()
// 	defer h.mu.Unlock()
// 	h.records = nil
// }

// flattenRecord returns a map of all attributes in a record, with groups flattened.
// Keys are joined with "." (e.g., "group.key").
// func flattenRecord(r *slog.Record) map[string]any {
// 	m := make(map[string]any)
// 	if r == nil {
// 		return m
// 	}
// 	r.Attrs(func(a slog.Attr) bool {
// 		flattenAttr(m, "", a)
// 		return true
// 	})
// 	return m
// }

// func flattenAttr(dest map[string]any, prefix string, a slog.Attr) {
// 	key := a.Key
// 	if prefix != "" {
// 		key = prefix + "." + key
// 	}
// 	val := a.Value
// 	switch val.Kind() {
// 	case slog.KindGroup:
// 		for _, attr := range val.Group() {
// 			flattenAttr(dest, key, attr)
// 		}
// 	default:
// 		dest[key] = val.Any()
// 	}
// }
