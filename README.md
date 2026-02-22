# gslog

A flexible logging package built on top of Go's `log/slog`, providing both **stateless** (plain `slog.Logger`) and **stateful** loggers that automatically enrich log records with fields from a user-defined state structure. It supports context‑based field extraction, reflection‑based field caching, and a clean functional options API.

---

## Features

- **Stateless logger** – plain `*slog.Logger` with configurable handlers (JSON/text), output, level, and options.
- **Stateful logger** – generic wrapper that adds non‑zero fields from a state struct to every log record.
  - Fields are grouped under the type name (or a custom group name).
  - Only exported fields are considered; nested structs are flattened with dot notation (`Address.City`).
  - Zero‑value fields can be optionally included.
- **Context enrichment** – extract values from `context.Context` and add them to log records.
  - Simple `ctx.Value(key)` or custom extractor functions.
  - Fields can be placed in a dedicated group.
- **Reflection caching** – field information is computed once per type and cached for performance.
- **Functional options** – configure loggers, handlers, and stateful behaviour with `ConfigOption` and `StatefulOption`.
- **Immutable updates** – `With`, `WithGroup`, `WithContextValue`, `UpdateState` return new logger instances, leaving the original unchanged.
- **Context propagation** – `EnrichContext` copies non‑zero state fields into a new context, making them available to other context‑aware components.
- **Two stability tracks**:
  - **Root (`github.com/Galdoba/gslog`)** – active development, API may change.
  - **v1 (`github.com/Galdoba/gslog/v1`)** – stable and sealed; no further changes.

---

## Installation

```bash
# Unstable development version (may break)
go get github.com/Galdoba/gslog

# Stable v1 (sealed, no further changes)
go get github.com/Galdoba/gslog/v1
```

---

## Quick Start

### Stateless Logger

```go
package main

import (
    "log/slog"
    "github.com/Galdoba/gslog"
)

func main() {
    cfg := gslog.NewSlogConfig(
        gslog.WithHandlerType("json"),
        gslog.WithLevel(slog.LevelDebug),
    )
    log := cfg.NewLogger()
    log.Info("hello world", "key", "value")
}
```

### Stateful Logger

```go
type User struct {
    ID    int
    Name  string
    Email string
}

func main() {
    cfg := gslog.NewSlogConfig(gslog.WithHandlerType("text"))
    state := &User{ID: 42, Name: "Alice", Email: "alice@example.com"}

    // Stateful logger adds non‑zero fields from *User to every log record.
    log := gslog.NewStateful(cfg, state,
        gslog.WithGroupName[User]("user"),          // optional custom group
        gslog.WithIncludeZeroFields[User](false),   // default
    )

    log.Info("user logged in")                       // includes {user.ID=42, user.Name=Alice, user.Email=...}
}
```

### Context Extraction

```go
ctx := context.WithValue(context.Background(), "request_id", "req-123")
ctx = context.WithValue(ctx, "trace_id", "trace-456")

fields := []gslog.ContextField{
    gslog.SimpleContextField("request_id"),
    gslog.ExtractorContextField("trace", func(ctx context.Context) any {
        return ctx.Value("trace_id")
    }),
}

log := gslog.NewLoggerWithContext(cfg, fields, "ctx") // fields grouped under "ctx"
log.InfoContext(ctx, "handling request")
// Output includes: ctx.request_id=req-123, ctx.trace=trace-456
```

### Enriching a Context from State

```go
state := &User{ID: 1, Name: "Bob"}
stateful := gslog.NewStateful(cfg, state)
ctx := gslog.EnrichContext(context.Background(), stateful)
// ctx now contains values for keys "ID", "Name"
```

---

## Why Stateful?

In many applications, certain data is part of the "context" of an operation – for example, a user ID, request ID, or tenant. Passing this data explicitly to every log call is tedious and error‑prone. A `Stateful` logger holds a pointer to a state structure and automatically includes its fields (excluding zero values) in every log record. This keeps your log calls clean while ensuring important context is never missing.

The state is stored as a pointer, so you can update it and create new logger instances with the updated state (`UpdateState`). All methods (`With`, `WithGroup`, etc.) return new loggers, making the API immutable by default.

---

## Stateless vs. Stateful

| Stateless                         | Stateful                                   |
|-----------------------------------|--------------------------------------------|
| Plain `*slog.Logger`              | Generic wrapper over `*slog.Logger`        |
| No automatic field injection      | Adds fields from a state struct automatically |
| Configured via `SlogConfig`       | Configured via `SlogConfig` + `StatefulOption` |
| Ideal for libraries or when no persistent context exists | Ideal for long‑lived components (handlers, services) with a well‑defined context |

---

## Configuration

`SlogConfig` controls the underlying `slog.Handler`:

```go
cfg := gslog.NewSlogConfig(
    gslog.WithHandlerType("json"),               // or "text"
    gslog.WithOutput(os.Stdout),
    gslog.WithLevel(slog.LevelWarn),
    gslog.WithHandlerOptions(&slog.HandlerOptions{AddSource: true}),
    gslog.WithCustomHandler(myHandler),          // overrides all above
)
```

---

## Stateful Options

- `WithGroupName[T](name string)` – place state fields under a custom group name.
- `WithIncludeZeroFields[T](include bool)` – include zero‑value fields (default `false`).

Example:
```go
log := gslog.NewStateful(cfg, state,
    gslog.WithGroupName[User]("user"),
    gslog.WithIncludeZeroFields[User](true),
)
```

---

## Context Extraction Handlers

Two mechanisms let you add context values to logs:

1. **`NewLoggerWithContext`** – creates a logger whose handler extracts a list of fields from the context on every call.
2. **`Stateful.WithContextValue`** – returns a new `Stateful` logger that extracts a single field using a custom extractor.

Both support grouping: if a group name is provided, all extracted fields are placed under that group.

---

## Reflection and Performance

Field information for a given type is computed once and cached using `sync.Map`. This makes repeated logging cheap – the reflection overhead is paid only the first time a type is used with a stateful logger.

---

## Versioning and Stability

- **Root (`github.com/Galdoba/gslog`)**  
  This is the active development branch. The API may change without notice as new features are added or refined. Use it if you need the latest enhancements and are willing to adapt to potential breaking changes.

- **v1 (`github.com/Galdoba/gslog/v1`)**  
  This version is **sealed**. It will receive no further changes – no new features, no bug fixes. It is provided for projects that require absolute stability and do not need ongoing updates. Import it as `github.com/Galdoba/gslog/v1`.

---

## Contributing

Issues and pull requests are welcome for the root version. For v1, only critical security fixes will be considered (and even then, a v1.0.1 might be released, but the API remains unchanged).

---

## Future: `slog` Reader App

A separate tool is planned that will:

- Ingest JSON logs from multiple sources.
- Format and print them as pretty, coloured text in the console.
- Support searching and filtering.

Stay tuned!
