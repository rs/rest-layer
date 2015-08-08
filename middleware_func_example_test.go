package rest_test

import (
	"net/http"

	"github.com/rs/rest-layer"
	"golang.org/x/net/context"
)

func validateCredentials(u, p string) bool {
	// auth logic
	return true
}

func ExampleNewMiddleware() {
	root := rest.New()
	api, _ := rest.NewHandler(root)

	// Add a very basic auth using a middleware
	api.Use(rest.NewMiddleware(func(ctx context.Context, r *http.Request, next rest.Next) (context.Context, int, http.Header, interface{}) {
		if u, p, ok := r.BasicAuth(); ok && validateCredentials(u, p) {
			// Store the authen user in the context
			ctx = context.WithValue(ctx, "user", u)
			// Pass to the next middleware
			return next(ctx)
		}
		// Stop the middleware chain and return a 401 HTTP error
		headers := http.Header{}
		headers.Set("WWW-Authenticate", "Basic realm=\"API\"")
		return ctx, 401, headers, &rest.Error{401, "Please provide proper credentials", nil}
	}))
}
