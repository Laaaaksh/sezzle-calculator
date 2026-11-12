// Package contextkeys owns the keys under which request-scoped values travel
// in a context.Context.
package contextkeys

import "context"

// key is unexported so no other package can build a colliding key. A plain
// string key would clash with any other package that picked the same text.
type key int

const (
	requestIDKey key = iota
)

// WithRequestID returns a child context carrying the request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID returns the request ID stored by WithRequestID. The bool is false
// when none was stored, so callers can tell "absent" from "empty".
func RequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}
