package rest

import (
	"log"
	"net/http"
	"strings"

	"github.com/rs/rest-layer/schema"
)

// Handler is an HTTP handler used to serve the configured REST API
type Handler struct {
	// ResponseSender can be changed to extend the DefaultResponseSender
	ResponseSender ResponseSender
	// resources stores the root map of resource -> resource definition
	resources map[string]*subResource
}

// New creates an new REST API HTTP handler with the specified validator
func New() *Handler {
	return &Handler{
		ResponseSender: DefaultResponseSender{},
		resources:      map[string]*subResource{},
	}
}

// Bind a resource at a specific route
func (h *Handler) Bind(name string, r *Resource) *Resource {
	// Compile schema and panic on any compilation error
	if c, ok := r.schema.(schema.Compiler); ok {
		if err := c.Compile(); err != nil {
			log.Fatalf("Schema compilation error: %s.%s", name, err)
		}
	}
	h.resources[name] = &subResource{resource: r}
	return r
}

// ServeHTTP handle requests as an http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &requestHandler{
		h:        h,
		req:      r,
		res:      w,
		s:        h.ResponseSender,
		skipBody: r.Method == "HEAD",
	}
	req.route(r.URL.Path, NewLookup(), h.resources)
}

// getResource retrives a given resource by it's path.
// For instance if a resource user has a sub-resource posts,
// a users.posts path can be use to retrieve the posts resource.
func (h *Handler) getResource(path string) *Resource {
	resources := h.resources
	var resource *Resource
	for _, comp := range strings.Split(path, ".") {
		if subResource, found := resources[comp]; found {
			resource = subResource.resource
			resources = resource.resources
		} else {
			return nil
		}
	}
	return resource
}
