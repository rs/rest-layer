package rest

import (
	"net/http"

	"golang.org/x/net/context"
)

// If is a middleware that evaluates a condition based on the current
// context and/or request and decides to execute on middleware or the
// other based on the condition's result.
type If struct {
	// Condition is evaluated at runtime to decide which middleware to
	// execute. If the Condition's function returns true, the middleware
	// stored in Then is executed otherwise Else is executed if defined.
	Condition func(ctx context.Context, r *http.Request) bool
	Then      Middleware
	Else      Middleware
}

// Handle makes the If middleware implement the Middleware interface
func (m If) Handle(ctx context.Context, r *http.Request, next Next) (context.Context, int, http.Header, interface{}) {
	// If no condition set, this middleware is just a pass thru
	if m.Condition != nil {
		if m.Condition(ctx, r) {
			if m.Then != nil {
				return m.Then.Handle(ctx, r, next)
			}
		} else {
			if m.Else != nil {
				return m.Else.Handle(ctx, r, next)
			}
		}
	}
	return next(ctx)
}
