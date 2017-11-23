package resource

import "context"

type ctxKey int

const (
	ctxKeyDisableHooks ctxKey = iota
)

// WithDisableHooks returns a new context where hooks are marked not to run.
func WithDisableHooks(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKeyDisableHooks, true)
}

func hooksDisabled(ctx context.Context) bool {
	b, ok := ctx.Value(ctxKeyDisableHooks).(bool)
	return ok && b
}
