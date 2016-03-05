package graphql

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/xlog"
	"golang.org/x/net/context"
)

// Handler is a net/http compatible handler used to serve the configured GraphQL API
type Handler struct {
	schema graphql.Schema
}

// NewHandler creates an new GraphQL API HTTP handler with the specified resource index
func NewHandler(i resource.Index) (*Handler, error) {
	if c, ok := i.(schema.Compiler); ok {
		if err := c.Compile(); err != nil {
			return nil, err
		}
	}
	// define schema, with our rootQuery and rootMutation
	s, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: newRootQuery(i),
	})
	if err != nil {
		return nil, err
	}
	return &Handler{schema: s}, nil
}

// getContext creates a context for the request to add net/context support when used as a
// standard http.Handler, without net/context support. The context will automatically be
// canceled as soon as passed request connection will be closed.
func getContext(w http.ResponseWriter, r *http.Request) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	// Handle canceled requests using net/context by passing a context
	// to the request handler that will be canceled as soon as the client
	// connection is closed
	if wcn, ok := w.(http.CloseNotifier); ok {
		notify := wcn.CloseNotify()
		go func() {
			// When client close the connection, cancel the context
			<-notify
			cancel()
		}()
	}
	return ctx
}

// ServeHTTP handles requests as a http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(w, r)
	h.ServeHTTPC(ctx, w, r)
}

// ServeHTTPC handles requests as a xhandler.HandlerC
func (h *Handler) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var query string
	switch r.Method {
	case "GET":
		query = r.URL.Query().Get("query")
	case "POST":
		b, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		query = string(b)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	result := graphql.Do(graphql.Params{
		Context:       ctx,
		RequestString: query,
		Schema:        h.schema,
	})
	if len(result.Errors) > 0 {
		xlog.FromContext(ctx).Errorf("wrong result, unexpected errors: %v", result.Errors)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
