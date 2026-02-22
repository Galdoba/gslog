package logger

import "context"

// ContextField defines a field to be extracted from context and added to log records.
type ContextField struct {
	Key       string
	Extractor func(context.Context) any // if nil, ctx.Value(Key) is used
}

// SimpleContextField creates a ContextField that uses ctx.Value(key) as the extractor.
// The field key is used as both the context key and the attribute key.
func SimpleContextField(key string) ContextField {
	return ContextField{Key: key}
}

// ExtractorContextField creates a ContextField with a custom extractor function.
// The key is used as the attribute key, and the extractor is called to obtain the value from context.
func ExtractorContextField(key string, extractor func(context.Context) any) ContextField {
	return ContextField{Key: key, Extractor: extractor}
}
