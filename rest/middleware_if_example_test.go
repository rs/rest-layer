package rest_test

import (
	"net/http"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"golang.org/x/net/context"
)

type SomeMiddleware struct{}

// Handle implements rest.Middleware interface
func (m *SomeMiddleware) Handle(ctx context.Context, r *http.Request, next rest.Next) (context.Context, int, http.Header, interface{}) {
	// code
	return next(ctx)
}

func ExampleIf() {
	index := resource.NewIndex()
	api, _ := rest.NewHandler(index)

	api.Use(rest.If{
		Condition: func(ctx context.Context, r *http.Request) bool {
			route, ok := rest.RouteFromContext(ctx)
			// True if current resource endpoint is users
			return ok && route.ResourcePath.Path() == "users"
		},
		Then: &SomeMiddleware{},
	})
}
