package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
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

// ServeHTTP handles requests as a http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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
		if r.Header.Get("Content-Type") == "application/json" {
			q := map[string]interface{}{}
			if err := json.Unmarshal(b, &q); err != nil {
				http.Error(w, fmt.Sprintf("Cannot unmarshal JSON: %v", err), http.StatusBadRequest)
			}
			query, _ = q["query"].(string)
		} else {
			query = string(b)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	result := graphql.Do(graphql.Params{
		Context:       ctx,
		RequestString: query,
		Schema:        h.schema,
	})
	if resource.Logger != nil {
		if len(result.Errors) > 0 {
			resource.Logger(ctx, resource.LogLevelError, fmt.Sprintf("wrong result, unexpected errors: %v", result.Errors), nil)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
