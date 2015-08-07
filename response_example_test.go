package rest_test

import (
	"net/http"

	"github.com/rs/rest-layer"
	"golang.org/x/net/context"
)

type myResponseSender struct {
	// Extending default response sender
	rest.DefaultResponseSender
}

// Add a wrapper around the list with pagination info
func (r myResponseSender) SendList(ctx context.Context, headers http.Header, l *rest.ItemList, skipBody bool) (context.Context, interface{}) {
	ctx, data := r.DefaultResponseSender.SendList(ctx, headers, l, skipBody)
	return ctx, map[string]interface{}{
		"meta": map[string]int{
			"total": l.Total,
			"page":  l.Page,
		},
		"list": data,
	}
}

func ExampleResponseSender() {
	root := rest.New()
	api, _ := rest.NewHandler(root)
	api.ResponseSender = myResponseSender{}
}
