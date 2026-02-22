package logger

import (
	"context"
	"log/slog"
	"reflect"
)

// Stateful is a generic wrapper around slog.Logger that automatically enriches
// log records with non-zero fields from a user-defined state structure of type T.
// The state is stored as a pointer and can be nil, in which case no fields are added.
// State fields are grouped under a key equal to the type name of T (or "state" if
// the type is unnamed), unless overridden by WithGroupName.
//
// Stateful[T] does not synchronize concurrent access to the state *T.
// If the state may be modified concurrently, external synchronization is required,
// or use immutable updates via UpdateState.
type Stateful[T any] struct {
	logger            *slog.Logger // underlying slog.Logger (never nil)
	state             *T           // pointer to the current state, may be nil
	groupName         string       // custom group name for state fields; if empty, derived from type
	includeZeroFields bool         // if true, zero-value fields are also logged

	config SlogConfig // configuration used to create this logger (may be zero if from external source)
}

// StatefulOption is a functional option for configuring a Stateful logger.
// These options affect only the Stateful wrapper, not the underlying slog.Logger.
type StatefulOption[T any] func(*Stateful[T])

// WithGroupName returns a StatefulOption that sets a custom group name for the state fields.
// If not set, the group name is derived from the type of T.
func WithGroupName[T any](name string) StatefulOption[T] {
	return func(l *Stateful[T]) {
		l.groupName = name
	}
}

// WithIncludeZeroFields returns a StatefulOption that controls whether zero-value fields
// are included in log records. By default, only non-zero fields are logged.
func WithIncludeZeroFields[T any](include bool) StatefulOption[T] {
	return func(l *Stateful[T]) {
		l.includeZeroFields = include
	}
}

// NewStateful creates a Stateful logger from the configuration with the given state
// and applies Stateful-specific options. The state may be nil.
func NewStateful[T any](c SlogConfig, state *T, opts ...StatefulOption[T]) *Stateful[T] {
	l := &Stateful[T]{
		logger:            c.NewLogger(),
		state:             state,
		groupName:         "",
		includeZeroFields: false,
		config:            c,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Modify applies the given StatefulOptions to a copy of the logger and returns the new instance.
// The original logger is not modified. This is useful for creating variations
// without affecting the original.
func Modify[T any](l *Stateful[T], opts ...StatefulOption[T]) *Stateful[T] {
	clone := &Stateful[T]{
		logger:            l.logger,
		state:             l.state,
		groupName:         l.groupName,
		includeZeroFields: l.includeZeroFields,
		config:            l.config,
	}
	for _, opt := range opts {
		opt(clone)
	}
	return clone
}

// WithState returns a new Stateful logger that uses the same underlying logger
// but with a different state type and value. The original logger is unchanged.
// This allows changing the state structure while keeping all other settings.
func WithState[T, U any](l *Stateful[T], state *U) *Stateful[U] {
	return &Stateful[U]{
		logger:            l.logger,
		state:             state,
		groupName:         "", // group name derived from new type U
		includeZeroFields: l.includeZeroFields,
		config:            l.config,
	}
}

// UpdateState returns a new Stateful logger with the same underlying logger and settings,
// but with the state replaced by the provided pointer. The original logger is unchanged.
func (l *Stateful[T]) UpdateState(state *T) *Stateful[T] {
	return &Stateful[T]{
		logger:            l.logger,
		state:             state,
		groupName:         l.groupName,
		includeZeroFields: l.includeZeroFields,
		config:            l.config,
	}
}

// stateTypeName returns the name to use for the state group.
// If a custom group name is set, it returns that; otherwise
// if the type of *l.state is named, it returns that name; otherwise "state".
func (l *Stateful[T]) stateTypeName() string {
	if l.groupName != "" {
		return l.groupName
	}
	if l.state == nil {
		return "state"
	}
	t := reflect.TypeOf(*l.state)
	if t.Name() != "" {
		return t.Name()
	}
	return "state"
}

// logWithState is an internal helper that adds state fields as a grouped attribute
// and delegates to the underlying slog.Logger.
func (l *Stateful[T]) logWithState(ctx context.Context, level slog.Level, msg string, args ...any) {
	if l.state == nil {
		l.logger.Log(ctx, level, msg, args...)
		return
	}
	stateAttrs := l.appendStateFields()
	if len(stateAttrs) == 0 {
		l.logger.Log(ctx, level, msg, args...)
		return
	}
	anyAttrs := make([]any, len(stateAttrs))
	for i, a := range stateAttrs {
		anyAttrs[i] = a
	}
	groupName := l.stateTypeName()
	group := slog.Group(groupName, anyAttrs...)
	finalArgs := append(args, group)
	l.logger.Log(ctx, level, msg, finalArgs...)
}

// appendStateFields returns a slice of slog.Attr for fields of the state.
// If includeZeroFields is false, only non-zero fields are included.
func (l *Stateful[T]) appendStateFields() []slog.Attr {
	if l.state == nil {
		return nil
	}
	t := reflect.TypeOf(*l.state)
	// Only structs have fields we can list.
	if t.Kind() != reflect.Struct {
		return nil
	}
	infos := getFieldInfos(t)
	if len(infos) == 0 {
		return nil
	}
	val := reflect.ValueOf(l.state).Elem()
	attrs := make([]slog.Attr, 0, len(infos))
	for _, fi := range infos {
		fval := val.FieldByIndex(fi.index)
		if l.includeZeroFields || !isZeroValue(fval) {
			attrs = append(attrs, slog.Any(fi.fullName, fval.Interface()))
		}
	}
	return attrs
}

// Debug logs at LevelDebug.
func (l *Stateful[T]) Debug(msg string, args ...any) {
	l.logWithState(context.Background(), slog.LevelDebug, msg, args...)
}

// Info logs at LevelInfo.
func (l *Stateful[T]) Info(msg string, args ...any) {
	l.logWithState(context.Background(), slog.LevelInfo, msg, args...)
}

// Warn logs at LevelWarn.
func (l *Stateful[T]) Warn(msg string, args ...any) {
	l.logWithState(context.Background(), slog.LevelWarn, msg, args...)
}

// Error logs at LevelError.
func (l *Stateful[T]) Error(msg string, args ...any) {
	l.logWithState(context.Background(), slog.LevelError, msg, args...)
}

// DebugContext logs at LevelDebug with the given context.
func (l *Stateful[T]) DebugContext(ctx context.Context, msg string, args ...any) {
	l.logWithState(ctx, slog.LevelDebug, msg, args...)
}

// InfoContext logs at LevelInfo with the given context.
func (l *Stateful[T]) InfoContext(ctx context.Context, msg string, args ...any) {
	l.logWithState(ctx, slog.LevelInfo, msg, args...)
}

// WarnContext logs at LevelWarn with the given context.
func (l *Stateful[T]) WarnContext(ctx context.Context, msg string, args ...any) {
	l.logWithState(ctx, slog.LevelWarn, msg, args...)
}

// ErrorContext logs at LevelError with the given context.
func (l *Stateful[T]) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.logWithState(ctx, slog.LevelError, msg, args...)
}

// With returns a new Stateful logger that includes the given additional attributes
// in every log record. The original logger is unchanged.
func (l *Stateful[T]) With(args ...any) *Stateful[T] {
	return &Stateful[T]{
		logger:            l.logger.With(args...),
		state:             l.state,
		groupName:         l.groupName,
		includeZeroFields: l.includeZeroFields,
		config:            l.config,
	}
}

// WithGroup returns a new Stateful logger that starts a group with the given name.
// All subsequent attributes (including state fields) will be placed under this group.
func (l *Stateful[T]) WithGroup(name string) *Stateful[T] {
	return &Stateful[T]{
		logger:            l.logger.WithGroup(name),
		state:             l.state,
		groupName:         l.groupName,
		includeZeroFields: l.includeZeroFields,
		config:            l.config,
	}
}

// WithContextValue returns a new Stateful logger that, when logging via context
// methods, will add a key-value pair extracted from the context.
// The original logger is unchanged.
func (l *Stateful[T]) WithContextValue(key string, extractor func(context.Context) any) *Stateful[T] {
	field := ContextField{Key: key, Extractor: extractor}
	handler := WrapHandlerWithContext(l.logger.Handler(), []ContextField{field}, "")
	return &Stateful[T]{
		logger:            slog.New(handler),
		state:             l.state,
		groupName:         l.groupName,
		includeZeroFields: l.includeZeroFields,
		config:            l.config,
	}
}

// Unwrap returns the underlying slog.Logger from a Stateful logger.
func Unwrap[T any](sl *Stateful[T]) *slog.Logger {
	return sl.logger
}

// MakeStateful creates a Stateful logger from an existing plain slog.Logger
// and a state pointer. The resulting logger will add non-zero fields from state
// to every log record, using the same underlying handler as the input logger.
// The config field is left zero because we cannot reconstruct the configuration.
func MakeStateful[T any](l *slog.Logger, state *T) *Stateful[T] {
	return &Stateful[T]{
		logger:            l,
		state:             state,
		groupName:         "",
		includeZeroFields: false,
		config:            SlogConfig{},
	}
}

// MakeStatefulWithContext creates a new Stateful logger from an existing plain slog.Logger
// and a state value. It creates a copy of the state, populates any zero fields from the
// context (using the full dotted field name as the context key), and returns a Stateful
// logger with a pointer to the enriched copy. The original state value is not modified.
func MakeStatefulWithContext[T any](ctx context.Context, l *slog.Logger, state T) *Stateful[T] {
	stateCopy := state
	t := reflect.TypeOf(state)
	// Only structs have fields we can populate from context.
	if t.Kind() == reflect.Struct {
		infos := getFieldInfos(t)
		if len(infos) > 0 {
			v := reflect.ValueOf(&stateCopy).Elem()
			for _, fi := range infos {
				fval := v.FieldByIndex(fi.index)
				if !isZeroValue(fval) {
					continue
				}
				if ctxVal := ctx.Value(fi.fullName); ctxVal != nil {
					rv := reflect.ValueOf(ctxVal)
					if rv.Type().AssignableTo(fval.Type()) {
						fval.Set(rv)
					} else if rv.Type().ConvertibleTo(fval.Type()) {
						fval.Set(rv.Convert(fval.Type()))
					}
				}
			}
		}
	}
	return &Stateful[T]{
		logger:            l,
		state:             &stateCopy,
		groupName:         "",
		includeZeroFields: false,
		config:            SlogConfig{},
	}
}

// EnrichContext returns a new context derived from parent, enriched with values
// from the state of the Stateful logger.
func EnrichContext[T any](parent context.Context, sl *Stateful[T]) context.Context {
	if sl.state == nil {
		return parent
	}
	t := reflect.TypeOf(*sl.state)
	// Only structs have fields we can extract.
	if t.Kind() != reflect.Struct {
		return parent
	}
	infos := getFieldInfos(t)
	if len(infos) == 0 {
		return parent
	}
	val := reflect.ValueOf(sl.state).Elem()
	ctx := parent
	for _, fi := range infos {
		fval := val.FieldByIndex(fi.index)
		if !isZeroValue(fval) {
			ctx = context.WithValue(ctx, fi.fullName, fval.Interface())
		}
	}
	return ctx
}

// Enabled reports whether the logger emits log records at the given level.
func (l *Stateful[T]) Enabled(ctx context.Context, level slog.Level) bool {
	return l.logger.Enabled(ctx, level)
}

// Handler returns the underlying handler of the logger.
func (l *Stateful[T]) Handler() slog.Handler {
	return l.logger.Handler()
}

// Log emits a log record at the given level, message and arguments.
// It automatically enriches the record with state fields (as a group).
func (l *Stateful[T]) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.logWithState(ctx, level, msg, args...)
}

// LogAttrs emits a log record at the given level, message and attributes.
// It converts the attributes to arguments and delegates to Log, which adds the state group.
func (l *Stateful[T]) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	// Convert []slog.Attr to []any where each element is the attr itself.
	// This works because slog.Logger's Log method treats Attr arguments specially.
	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}
	l.Log(ctx, level, msg, args...)
}
