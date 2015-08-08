package rest

import (
	"net/http"

	"golang.org/x/net/context"
)

// Middleware are called after routing has been resolved and before the found route is processed.
//
// A middleware may access the found route using rest.RouteFromContext(ctx) or the router itself
// using rest.RouterFromContext(ctx).
//
// A middleware can choose to either directly answer to skip the other handlers or to pass
// to the next handler by calling "next".
//
// A middleware returns an eventually modified context, the HTTP status, a list of HTTP headers
// to append to the response and a response body. The response body may be any kind of object
// the Response sender is able to handle. The default response sender can handle rest.Item,
// rest.ListItem, rest.Error, error or any JSON serializable type.
type Middleware interface {
	Handle(ctx context.Context, r *http.Request, next Next) (context.Context, int, http.Header, interface{})
}

// Next is the callback handler called by middelware to pass the the next handler
type Next func(ctx context.Context) (context.Context, int, http.Header, interface{})

// middlewareFuncWrapper is used to wrap a middleware handler function in order to
// comply with the middleware interface
type middlewareFuncWrapper struct {
	handleFunc func(ctx context.Context, r *http.Request, next Next) (context.Context, int, http.Header, interface{})
}

func (m middlewareFuncWrapper) Handle(ctx context.Context, r *http.Request, next Next) (context.Context, int, http.Header, interface{}) {
	return m.handleFunc(ctx, r, next)
}

// Use adds a middleware the the middleware chain
//
// WARNING: this method is not thread safe. You should never add a middleware while
// the http.Handler is serving requests.
func (h *Handler) Use(m Middleware) {
	h.mw = append(h.mw, m)
}

// UseFunc adds a middleware the the middleware chain as a function
//
// WARNING: this method is not thread safe. You should never add a middleware while
// the http.Handler is serving requests.
func (h *Handler) UseFunc(f func(ctx context.Context, r *http.Request, next Next) (context.Context, int, http.Header, interface{})) {
	h.mw = append(h.mw, &middlewareFuncWrapper{f})
}

func (h *Handler) callMiddlewares(ctx context.Context, r *http.Request, last Next) (context.Context, int, http.Header, interface{}) {
	l := len(h.mw)
	if l == 0 {
		return last(ctx)
	}
	i := -1
	var next Next
	next = func(ctx context.Context) (context.Context, int, http.Header, interface{}) {
		i++
		if i < l {
			return h.mw[i].Handle(ctx, r, next)
		}
		return last(ctx)
	}
	return next(ctx)
}
