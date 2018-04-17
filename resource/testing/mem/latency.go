package mem

import (
	"context"
	"time"
)

// handleWithLatency allows introduction of artificial latency while handling context cancellation.
// The method first wait for the given latency while monitoring ctx.Done. If context is canceled
// during the wait, the context error is returned.
// If latency passed, the handler is executed and it's error output is returned.
func handleWithLatency(latency time.Duration, ctx context.Context, handler func() error) error {
	if latency == 0 {
		return handler()
	}

	select {
	case <-ctx.Done():
		// Monitor context cancellation. cancellation may happend if the client closed the connection
		// or if the configured request timeout has been reached.
		return ctx.Err()
	case <-time.After(latency):
		// Wait for the given latency before the execute the provided handler.
		return handler()
	}
}
