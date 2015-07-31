package rest

import "net/http"

// Handler is an HTTP handler used to serve the configured REST API
type Handler struct {
	// ResponseSender can be changed to extend the DefaultResponseSender
	ResponseSender ResponseSender
	// root stores the root resource
	root *RootResource
}

// NewHandler creates an new REST API HTTP handler with the specified root resource
func NewHandler(r *RootResource) (*Handler, error) {
	if err := r.Compile(); err != nil {
		return nil, err
	}
	h := &Handler{
		ResponseSender: DefaultResponseSender{},
		root:           r,
	}
	return h, nil
}

// ServeHTTP handle requests as an http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &requestHandler{
		root:     h.root,
		req:      r,
		res:      w,
		s:        h.ResponseSender,
		skipBody: r.Method == "HEAD",
	}
	req.route(r.URL.Path, NewLookup(), h.root.resources)
}
