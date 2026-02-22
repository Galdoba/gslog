package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

type testPerson struct {
	Name    string
	Age     int
	Address struct {
		City string
		Zip  int
	}
}

func TestStatefulLogging(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	state := &testPerson{
		Name: "Alice",
		Age:  30,
	}
	state.Address.City = "Paris"
	state.Address.Zip = 75001

	logger := NewStateful(cfg, state)
	logger.Info("hello")

	rec := th.lastRecord()
	if rec == nil {
		t.Fatal("no record")
	}
	attrs := flattenRecord(rec)

	if attrs["testPerson.Name"] != "Alice" {
		t.Errorf("testPerson.Name = %v, want Alice", attrs["testPerson.Name"])
	}
	if attrs["testPerson.Age"] != int64(30) {
		t.Errorf("testPerson.Age = %v, want 30", attrs["testPerson.Age"])
	}
	if attrs["testPerson.Address.City"] != "Paris" {
		t.Errorf("testPerson.Address.City = %v, want Paris", attrs["testPerson.Address.City"])
	}
	if attrs["testPerson.Address.Zip"] != int64(75001) {
		t.Errorf("testPerson.Address.Zip = %v, want 75001", attrs["testPerson.Address.Zip"])
	}
}

func TestStatefulWithGroupName(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	state := &testPerson{Name: "Bob"}

	logger := NewStateful(cfg, state, WithGroupName[testPerson]("custom"))
	logger.Info("msg")

	rec := th.lastRecord()
	attrs := flattenRecord(rec)
	if attrs["custom.Name"] != "Bob" {
		t.Errorf("custom.Name = %v, want Bob", attrs["custom.Name"])
	}
}

func TestStatefulIncludeZeroFields(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	state := &testPerson{Name: "Charlie"} // Age and Address zero

	t.Run("exclude zero (default)", func(t *testing.T) {
		th.reset()
		logger := NewStateful(cfg, state)
		logger.Info("msg")
		rec := th.lastRecord()
		attrs := flattenRecord(rec)
		if _, ok := attrs["testPerson.Age"]; ok {
			t.Error("Age field should be omitted")
		}
	})

	t.Run("include zero", func(t *testing.T) {
		th.reset()
		logger := NewStateful(cfg, state, WithIncludeZeroFields[testPerson](true))
		logger.Info("msg")
		rec := th.lastRecord()
		attrs := flattenRecord(rec)
		if attrs["testPerson.Age"] != int64(0) {
			t.Errorf("Age = %v, want 0", attrs["testPerson.Age"])
		}
		if attrs["testPerson.Address.City"] != "" {
			t.Errorf("Address.City = %v, want empty", attrs["testPerson.Address.City"])
		}
		if attrs["testPerson.Address.Zip"] != int64(0) {
			t.Errorf("Address.Zip = %v, want 0", attrs["testPerson.Address.Zip"])
		}
	})
}

func TestStatefulWithState(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	state1 := &testPerson{Name: "Alice"}
	logger1 := NewStateful(cfg, state1)

	state2 := "some string"
	logger2 := WithState(logger1, &state2)

	logger2.Info("msg")
	rec := th.lastRecord()
	attrs := flattenRecord(rec)
	// T is string, which has no exported fields -> no state attributes.
	if len(attrs) != 0 {
		t.Errorf("expected no attributes, got %v", attrs)
	}
}

func TestStatefulUpdateState(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	state := &testPerson{Name: "Alice"}
	logger := NewStateful(cfg, state)

	logger.Info("first")
	attrs1 := flattenRecord(th.lastRecord())
	if attrs1["testPerson.Name"] != "Alice" {
		t.Error("first log wrong")
	}

	newState := &testPerson{Name: "Bob"}
	logger2 := logger.UpdateState(newState)
	logger2.Info("second")
	attrs2 := flattenRecord(th.lastRecord())
	if attrs2["testPerson.Name"] != "Bob" {
		t.Errorf("after update Name = %v, want Bob", attrs2["testPerson.Name"])
	}

	// Original logger unchanged
	logger.Info("third")
	attrs3 := flattenRecord(th.lastRecord())
	if attrs3["testPerson.Name"] != "Alice" {
		t.Error("original logger changed")
	}
}

func TestStatefulWithContextValue(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	state := &testPerson{Name: "Alice"}
	logger := NewStateful(cfg, state)
	ctx := context.WithValue(context.Background(), "trace_id", "123")

	logger2 := logger.WithContextValue("trace_id", func(ctx context.Context) any {
		return ctx.Value("trace_id")
	})
	logger2.InfoContext(ctx, "msg")

	rec := th.lastRecord()
	attrs := flattenRecord(rec)
	if attrs["testPerson.Name"] != "Alice" {
		t.Error("state missing")
	}
	if attrs["trace_id"] != "123" {
		t.Errorf("trace_id = %v, want 123", attrs["trace_id"])
	}
}

func TestStatefulMakeStatefulWithContext(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	baseLogger := cfg.NewLogger()
	state := testPerson{} // all zero

	ctx := context.Background()
	ctx = context.WithValue(ctx, "Name", "Alice")
	ctx = context.WithValue(ctx, "Address.City", "Paris")

	logger := MakeStatefulWithContext(ctx, baseLogger, state)
	logger.Info("msg")

	rec := th.lastRecord()
	attrs := flattenRecord(rec)
	if attrs["testPerson.Name"] != "Alice" {
		t.Errorf("Name = %v, want Alice", attrs["testPerson.Name"])
	}
	if attrs["testPerson.Address.City"] != "Paris" {
		t.Errorf("Address.City = %v, want Paris", attrs["testPerson.Address.City"])
	}
	// Age and Zip should be zero and omitted (includeZero false)
	if _, ok := attrs["testPerson.Age"]; ok {
		t.Error("Age should be omitted")
	}
	if _, ok := attrs["testPerson.Address.Zip"]; ok {
		t.Error("Zip should be omitted")
	}
}

func TestStatefulEnrichContext(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	state := &testPerson{
		Name: "Alice",
		Age:  30,
	}
	state.Address.City = "Paris"
	state.Address.Zip = 75001

	logger := NewStateful(cfg, state)

	parent := context.Background()
	ctx := EnrichContext(parent, logger)

	if val := ctx.Value("Name"); val != "Alice" {
		t.Errorf("Name = %v, want Alice", val)
	}
	if val := ctx.Value("Age"); val != 30 {
		t.Errorf("Age = %v, want 30", val)
	}
	if val := ctx.Value("Address.City"); val != "Paris" {
		t.Errorf("Address.City = %v, want Paris", val)
	}
	if val := ctx.Value("Address.Zip"); val != 75001 {
		t.Errorf("Address.Zip = %v, want 75001", val)
	}
}

func TestStatefulWithGroup(t *testing.T) {
	var buf bytes.Buffer
	cfg := NewSlogConfig(
		WithHandlerType("json"),
		WithOutput(&buf),
	)
	state := &testPerson{Name: "Alice"}
	logger := NewStateful(cfg, state).WithGroup("app")
	logger.Info("msg")

	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	appGroup, ok := record["app"].(map[string]any)
	if !ok {
		t.Fatal("expected 'app' group at top level")
	}
	stateGroup, ok := appGroup["testPerson"].(map[string]any)
	if !ok {
		t.Fatal("expected 'testPerson' group inside 'app'")
	}
	if name, ok := stateGroup["Name"].(string); !ok || name != "Alice" {
		t.Errorf("Name = %v, want Alice", stateGroup["Name"])
	}
}
func TestStatefulWith(t *testing.T) {
	var buf bytes.Buffer
	cfg := NewSlogConfig(
		WithHandlerType("json"),
		WithOutput(&buf),
	)
	state := &testPerson{Name: "Alice"}
	logger := NewStateful(cfg, state).With("extra", "value")
	logger.Info("msg")

	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Check that the extra attribute is present at top level
	if extra, ok := record["extra"].(string); !ok || extra != "value" {
		t.Errorf("extra = %v, want 'value'", record["extra"])
	}
	// Check that state is still present (inside testPerson group)
	stateGroup, ok := record["testPerson"].(map[string]any)
	if !ok {
		t.Fatal("expected 'testPerson' group")
	}
	if name, ok := stateGroup["Name"].(string); !ok || name != "Alice" {
		t.Errorf("Name = %v, want Alice", stateGroup["Name"])
	}
}

// testStateStruct is a simple struct for stateful tests.
type testStateStruct struct {
	Name string
	Age  int
}

func TestStatefulEnabled(t *testing.T) {
	// Create a handler that enables only levels >= Warn
	th := &enabledTestHandler{minLevel: slog.LevelWarn}
	cfg := NewSlogConfig(WithCustomHandler(th))
	state := &testStateStruct{Name: "Alice", Age: 30}
	sl := NewStateful(cfg, state)

	ctx := context.Background()
	if sl.Enabled(ctx, slog.LevelDebug) {
		t.Error("Enabled(LevelDebug) = true, want false")
	}
	if sl.Enabled(ctx, slog.LevelInfo) {
		t.Error("Enabled(LevelInfo) = true, want false")
	}
	if !sl.Enabled(ctx, slog.LevelWarn) {
		t.Error("Enabled(LevelWarn) = false, want true")
	}
	if !sl.Enabled(ctx, slog.LevelError) {
		t.Error("Enabled(LevelError) = false, want true")
	}
}

// enabledTestHandler is a handler that reports Enabled based on a minLevel.
type enabledTestHandler struct {
	minLevel slog.Level
	records  []slog.Record
}

func (h *enabledTestHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.minLevel
}
func (h *enabledTestHandler) Handle(ctx context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}
func (h *enabledTestHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *enabledTestHandler) WithGroup(name string) slog.Handler       { return h }

func TestStatefulHandler(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	state := &testStateStruct{Name: "Alice"}
	sl := NewStateful(cfg, state)

	// Handler should return the handler from the underlying logger.
	// Since the underlying logger uses th, we can compare via pointer.
	gotHandler := sl.Handler()
	if gotHandler != th {
		t.Errorf("Handler() returned %v, want %v", gotHandler, th)
	}
}

func TestStatefulLog(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	state := &testStateStruct{Name: "Alice", Age: 30}
	sl := NewStateful(cfg, state)

	ctx := context.Background()
	sl.Log(ctx, slog.LevelInfo, "test message", "extra", "value")

	rec := th.lastRecord()
	if rec == nil {
		t.Fatal("no record captured")
	}
	if rec.Level != slog.LevelInfo {
		t.Errorf("level = %v, want %v", rec.Level, slog.LevelInfo)
	}
	if rec.Message != "test message" {
		t.Errorf("message = %q, want %q", rec.Message, "test message")
	}

	attrs := flattenRecord(rec)
	// State fields should be present under the type name ("testStateStruct")
	if attrs["testStateStruct.Name"] != "Alice" {
		t.Errorf("Name = %v, want Alice", attrs["testStateStruct.Name"])
	}
	if attrs["testStateStruct.Age"] != int64(30) {
		t.Errorf("Age = %v, want 30", attrs["testStateStruct.Age"])
	}
	// Extra argument should be present
	if attrs["extra"] != "value" {
		t.Errorf("extra = %v, want value", attrs["extra"])
	}
}

func TestStatefulLogWithContext(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	state := &testStateStruct{Name: "Alice"}
	sl := NewStateful(cfg, state)

	ctx := context.WithValue(context.Background(), "req_id", "123")
	sl.Log(ctx, slog.LevelInfo, "msg")

	rec := th.lastRecord()
	if rec == nil {
		t.Fatal("no record")
	}
	// The context itself is not automatically added; only if the handler extracts it.
	// So we just verify state is still there.
	attrs := flattenRecord(rec)
	if attrs["testStateStruct.Name"] != "Alice" {
		t.Errorf("Name = %v, want Alice", attrs["testStateStruct.Name"])
	}
}

func TestStatefulLogAttrs(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	state := &testStateStruct{Name: "Alice", Age: 30}
	sl := NewStateful(cfg, state)

	ctx := context.Background()
	attrs := []slog.Attr{
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	}
	sl.LogAttrs(ctx, slog.LevelInfo, "test with attrs", attrs...)

	rec := th.lastRecord()
	if rec == nil {
		t.Fatal("no record")
	}
	flat := flattenRecord(rec)
	if flat["testStateStruct.Name"] != "Alice" {
		t.Errorf("Name = %v, want Alice", flat["testStateStruct.Name"])
	}
	if flat["testStateStruct.Age"] != int64(30) {
		t.Errorf("Age = %v, want 30", flat["testStateStruct.Age"])
	}
	if flat["key1"] != "value1" {
		t.Errorf("key1 = %v, want value1", flat["key1"])
	}
	if flat["key2"] != int64(42) {
		t.Errorf("key2 = %v, want 42", flat["key2"])
	}
}

func TestStatefulLogWithGroup(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	state := &testStateStruct{Name: "Alice"}
	sl := NewStateful(cfg, state, WithGroupName[testStateStruct]("custom"))

	ctx := context.Background()
	sl.Log(ctx, slog.LevelInfo, "msg")

	rec := th.lastRecord()
	flat := flattenRecord(rec)
	if flat["custom.Name"] != "Alice" {
		t.Errorf("custom.Name = %v, want Alice", flat["custom.Name"])
	}
}

func TestStatefulLogIncludeZeroFields(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	state := &testStateStruct{Name: "Alice"} // Age is zero

	t.Run("exclude zero (default)", func(t *testing.T) {
		th.reset()
		sl := NewStateful(cfg, state)
		sl.Log(context.Background(), slog.LevelInfo, "msg")
		flat := flattenRecord(th.lastRecord())
		if _, ok := flat["testStateStruct.Age"]; ok {
			t.Error("Age field should be omitted")
		}
	})

	t.Run("include zero", func(t *testing.T) {
		th.reset()
		sl := NewStateful(cfg, state, WithIncludeZeroFields[testStateStruct](true))
		sl.Log(context.Background(), slog.LevelInfo, "msg")
		flat := flattenRecord(th.lastRecord())
		if flat["testStateStruct.Age"] != int64(0) {
			t.Errorf("Age = %v, want 0", flat["testStateStruct.Age"])
		}
	})
}

func TestStatefulLogNilState(t *testing.T) {
	th := newTestHandler()
	cfg := NewSlogConfig(WithCustomHandler(th))
	sl := NewStateful[testStateStruct](cfg, nil)

	ctx := context.Background()
	sl.Log(ctx, slog.LevelInfo, "msg", "extra", "value")

	rec := th.lastRecord()
	if rec == nil {
		t.Fatal("no record")
	}
	flat := flattenRecord(rec)
	// No state group should appear
	if _, ok := flat["testStateStruct"]; ok {
		t.Error("state group should not appear when state is nil")
	}
	if flat["extra"] != "value" {
		t.Errorf("extra = %v, want value", flat["extra"])
	}
}

func TestStatefulLogAttrsWithGroup(t *testing.T) {
	var buf bytes.Buffer
	cfg := NewSlogConfig(
		WithHandlerType("json"),
		WithOutput(&buf),
	)
	state := &testStateStruct{Name: "Alice"}
	sl := NewStateful(cfg, state).WithGroup("app")

	ctx := context.Background()
	sl.LogAttrs(ctx, slog.LevelInfo, "msg", slog.String("request_id", "123"))

	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Check top-level group "app"
	appGroup, ok := record["app"].(map[string]any)
	if !ok {
		t.Fatal("expected 'app' group at top level")
	}

	// Inside "app", there should be a group named after the state type ("testStateStruct")
	stateGroup, ok := appGroup["testStateStruct"].(map[string]any)
	if !ok {
		t.Fatal("expected 'testStateStruct' group inside 'app'")
	}
	if name, ok := stateGroup["Name"].(string); !ok || name != "Alice" {
		t.Errorf("stateGroup['Name'] = %v, want 'Alice'", stateGroup["Name"])
	}

	// Also inside "app", there should be the "request_id" attribute
	if reqID, ok := appGroup["request_id"].(string); !ok || reqID != "123" {
		t.Errorf("appGroup['request_id'] = %v, want '123'", appGroup["request_id"])
	}
}
