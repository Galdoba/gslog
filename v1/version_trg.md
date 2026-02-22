## gslog/v1: Technical Reference Guide

**Version:** v1 (Sealed)
**Import Path:** `github.com/Galdoba/gslog/v1`

This document provides a comprehensive reference for the `v1` release of the `gslog` library. This version is **sealed**, meaning its API is frozen and will not receive any further changes, feature additions, or bug fixes. It is intended for projects that require absolute stability.

### 1. Overview

`gslog/v1` is a flexible logging package built on top of Go's standard `log/slog` package. It provides two primary logging paradigms:

*   **Stateless Logger**: A plain `*slog.Logger` configured via the `SlogConfig` builder. This is suitable for simple logging needs or for libraries that should not carry implicit state.
*   **Stateful Logger**: A generic wrapper (`Stateful[T]`) that automatically enriches every log record with fields from a user-defined state structure (e.g., a struct representing a user, request, or service context). Fields are grouped, and only exported fields are considered.

Key features include:
*   **Context Enrichment**: Extract values from `context.Context` and add them to log records using `ContextExtractorHandler`.
*   **Reflection Caching**: Field information for stateful loggers is computed once per type and cached, minimizing runtime overhead.
*   **Immutable API**: Methods like `With`, `WithGroup`, and `UpdateState` return new logger instances, leaving the original unchanged.
*   **Configuration**: A functional options pattern for configuring both the underlying `slog.Handler` and the stateful wrapper.

### 2. Core Configuration: `SlogConfig`

The foundation of any logger is its configuration. `SlogConfig` holds all parameters needed to create an `slog.Handler`.

```go
type SlogConfig struct {
    HandlerType    string                // "json" or "text". Defaults to "json".
    Output         io.Writer              // Destination. Defaults to os.Stderr.
    Level          slog.Leveler           // Minimum level. Defaults to slog.LevelInfo.
    HandlerOptions *slog.HandlerOptions   // e.g., AddSource, ReplaceAttr.
    CustomHandler  slog.Handler           // If non-nil, overrides all other fields.
}
```

#### 2.1. Creating a Config with Options

Use `NewSlogConfig` with `ConfigOption` functions.

```go
import "github.com/Galdoba/gslog/v1"

func main() {
    cfg := logger.NewSlogConfig(
        logger.WithHandlerType("text"),
        logger.WithLevel(slog.LevelDebug),
        logger.WithHandlerOptions(&slog.HandlerOptions{AddSource: true}),
        // logger.WithOutput(os.Stdout),
        // logger.WithCustomHandler(myCustomHandler),
    )
}
```

**Config Options:**

*   `func WithHandlerType(handlerType string) ConfigOption`
*   `func WithOutput(w io.Writer) ConfigOption`
*   `func WithLevel(level slog.Level) ConfigOption`
*   `func WithHandlerOptions(opts *slog.HandlerOptions) ConfigOption`
*   `func WithCustomHandler(handler slog.Handler) ConfigOption`

The `SlogConfigDefault()` function returns a config matching the default `slog` logger: a text handler to stderr with `Info` level.

### 3. Stateless Loggers

A stateless logger is a plain `*slog.Logger`. It is created directly from a `SlogConfig`.

```go
cfg := logger.NewSlogConfig(logger.WithHandlerType("json"))
log := cfg.NewLogger() // or cfg.Stateless()

log.Info("server starting", "port", 8080)
```

#### 3.1. Context-Aware Stateless Logger

`NewLoggerWithContext` creates a logger whose handler automatically extracts specified fields from the `context.Context` on every logging call.

```go
fields := []logger.ContextField{
    logger.SimpleContextField("request_id"),
    logger.ExtractorContextField("trace_id", func(ctx context.Context) any {
        // Custom extraction logic
        return ctx.Value("trace-id-key")
    }),
}

log := logger.NewLoggerWithContext(cfg, fields, "http") // Fields grouped under "http"

// Later, in an HTTP handler:
ctx := context.WithValue(context.Background(), "request_id", "req-123")
log.InfoContext(ctx, "handling request")
// Output includes: http.request_id=req-123, and the extracted trace_id.
```

**`ContextField` Constructors:**

*   `func SimpleContextField(key string) ContextField`: Uses `ctx.Value(key)` to get the value.
*   `func ExtractorContextField(key string, extractor func(context.Context) any) ContextField`: Uses a custom function to extract the value. The `key` is used as the log attribute key.

### 4. Stateful Loggers

The stateful logger, `Stateful[T]`, is the core feature of this package. It automatically adds fields from a state object to every log record.

```go
type AppState struct {
    UserID    int
    RequestID string
    Version   string
}

func main() {
    cfg := logger.NewSlogConfig(logger.WithHandlerType("text"))
    state := &AppState{UserID: 123, Version: "1.0"}

    log := logger.NewStateful(cfg, state,
        logger.WithGroupName[AppState]("app"),          // Optional: custom group name
        logger.WithIncludeZeroFields[AppState](false), // Optional: exclude zero values (default)
    )

    log.Info("user action") 
    // Output includes: app.UserID=123 app.Version="1.0" (RequestID is zero and omitted)
}
```

#### 4.1. Type Parameters and Options

`Stateful[T any]` is generic over the type of your state structure.

**Stateful-Specific Options (`StatefulOption[T]`):**

*   `func WithGroupName[T any](name string) StatefulOption[T]`: Places all state fields under a custom group name. Defaults to the name of type `T` (or "state" if the type is unnamed).
*   `func WithIncludeZeroFields[T any](include bool) StatefulOption[T]`: If `true`, zero-value fields are also logged. Default is `false` (zero fields omitted).

#### 4.2. Core Methods

`Stateful[T]` mirrors the `slog.Logger` methods, both with and without `context.Context`. All methods are immutable; they return a new `Stateful` instance.

*   **Logging Methods:**
    *   `Debug(msg string, args ...any)`
    *   `Info(msg string, args ...any)`
    *   `Warn(msg string, args ...any)`
    *   `Error(msg string, args ...any)`
    *   `DebugContext(ctx context.Context, msg string, args ...any)`
    *   `InfoContext(ctx context.Context, msg string, args ...any)`
    *   `WarnContext(ctx context.Context, msg string, args ...any)`
    *   `ErrorContext(ctx context.Context, msg string, args ...any)`

*   **Builder Methods:**
    *   `func (l *Stateful[T]) With(args ...any) *Stateful[T]`: Returns a new logger with additional attributes.
    *   `func (l *Stateful[T]) WithGroup(name string) *Stateful[T]`: Returns a new logger that starts a group.
    *   `func (l *Stateful[T]) WithContextValue(key string, extractor func(context.Context) any) *Stateful[T]`: Returns a new logger that extracts a single value from the context on every call.

*   **State Management:**
    *   `func (l *Stateful[T]) UpdateState(state *T) *Stateful[T]`: Returns a new logger with the same configuration but a different state pointer.
    *   `func WithState[T, U any](l *Stateful[T], state *U) *Stateful[U]`: Creates a new stateful logger with a different state type `U`, reusing the underlying `slog.Logger`.

#### 4.3. Creating from an Existing `slog.Logger`

*   `func MakeStateful[T any](l *slog.Logger, state *T) *Stateful[T]`: Wraps an existing plain `*slog.Logger` to make it stateful.
*   `func MakeStatefulWithContext[T any](ctx context.Context, l *slog.Logger, state T) *Stateful[T]`: Creates a copy of `state`, populates any of its zero fields with values from `ctx` (using the field's full dotted name as the context key), and returns a stateful logger with a pointer to the enriched copy.

    ```go
    // Assuming a context with values for keys "UserID" and "Address.City"
    statefulLog := logger.MakeStatefulWithContext(ctx, slog.Default(), AppState{})
    ```

### 5. Reflection and Field Handling

Stateful loggers rely on reflection to inspect state structs. This information is cached for performance.

*   **Field Discovery**: All exported fields of a struct are discovered. Nested structs are flattened using dot notation (e.g., `Address.City`).
*   **Zero Value Exclusion**: By default, fields with zero values (e.g., `0`, `""`, `nil`, `false`) are omitted from log records. This can be overridden with `WithIncludeZeroFields`.
*   **Caching**: The `getFieldInfos` function caches the field metadata per type, so the reflection overhead is paid only once per type during the application's lifetime.

### 6. Context Enrichment and Extraction

The library provides bi-directional integration between state and context.

*   **From Context to Logger**: `WithContextValue` (on `Stateful`) and `NewLoggerWithContext` (stateless) allow loggers to extract values from the context on each log call.
*   **From State to Context**: `func EnrichContext[T any](parent context.Context, sl *Stateful[T]) context.Context` returns a new context derived from `parent`, enriched with the non-zero fields from the stateful logger's state. The context keys are the full dotted field names.

    ```go
    ctx := logger.EnrichContext(context.Background(), myStatefulLog)
    // Later, any component with access to ctx can retrieve values:
    userID := ctx.Value("UserID")
    ```

### 7. Underlying Handler: `ContextExtractorHandler`

This handler (`handlers.go`) is responsible for adding context fields. It can be used independently, though the package's helper functions are the recommended way to create it.

```go
handler := logger.WrapHandlerWithContext(baseHandler, fields, "groupName")
```

It correctly propagates `WithAttrs` and `WithGroup` calls to the wrapped handler.

### 8. Sealed Version Guarantees (v1)

This version, `github.com/Galdoba/gslog/v1`, is **sealed**. The following guarantees are made:
*   **No API Changes**: No new functions, methods, types, or constants will be added.
*   **No Behavioral Changes**: The existing behavior of all public APIs will not be modified.
*   **No Bug Fixes**: Even if a bug is discovered, it will not be fixed in this version. Users requiring fixes must transition to the root development version (`github.com/Galdoba/gslog`), which may contain breaking changes.
*   **Import Path Stability**: The import path `github.com/Galdoba/gslog/v1` will always point to this exact API.

This stability makes it suitable for vendoring and long-term projects where dependency changes must be minimized.
