package rest_test

import (
	"net/http"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"golang.org/x/net/context"
)

type AuthMiddleware struct {
	Authenticator func(user string, password string) bool
}

func (m *AuthMiddleware) Handle(ctx context.Context, r *http.Request, next rest.Next) (context.Context, int, http.Header, interface{}) {
	if u, p, ok := r.BasicAuth(); ok && m.Authenticator(u, p) {
		// Store the authen user in the context
		ctx = context.WithValue(ctx, "user", u)
		// Pass to the next middleware
		return next(ctx)
	}
	// Stop the middleware chain and return a 401 HTTP error
	headers := http.Header{}
	headers.Set("WWW-Authenticate", "Basic realm=\"API\"")
	return ctx, 401, headers, &rest.Error{401, "Please provide proper credentials", nil}

}

func ExampleMiddleware() {
	index := resource.NewIndex()
	api, _ := rest.NewHandler(index)

	// Add a very basic auth using a middleware
	api.Use(&AuthMiddleware{
		Authenticator: func(user string, password string) bool {
			// code to check credentials
			return true
		},
	})
}
