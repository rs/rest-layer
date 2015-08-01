package mem

import (
	"time"

	"github.com/rs/rest-layer"
	"golang.org/x/net/context"
)

// handleWithLatency allows introduction of artificial latency while handling context cancellation.
// The method first wait for the given latency while monitoring ctx.Done. If context is canceled
// during the wait, the context error is returned (after being wrapped into rest.ContextError()).
// If latency passed, the handler is executed and it's error output is returned.
func handleWithLatency(latency time.Duration, ctx context.Context, handler func() *rest.Error) *rest.Error {
	select {
	case <-ctx.Done():
		// Monitor context cancellation. cancellation may happend if the client closed the connection
		// or if the configured request timeout has been reached.
		return rest.ContextError(ctx.Err())
	case <-time.After(latency):
		// Wait for the given latency before the execute the provided handler.
		return handler()
	}
}
