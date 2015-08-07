package rest

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"
)

// Handler is an HTTP handler used to serve the configured REST API
type Handler struct {
	// ResponseSender can be changed to extend the DefaultResponseSender
	ResponseSender ResponseSender
	// RequestTimeout is the default timeout for requests after which the whole request
	// is abandonned. The default value is no timeout.
	RequestTimeout time.Duration
	// router stores the resource router
	router ResourceRouter
}

// NewHandler creates an new REST API HTTP handler with the specified root resource
func NewHandler(r *RootResource) (*Handler, error) {
	if err := r.Compile(); err != nil {
		return nil, err
	}
	h := &Handler{
		ResponseSender: DefaultResponseSender{},
		router:         r,
	}
	return h, nil
}

// getTimeout get request timeout info from request or server config
func (h *Handler) getTimeout(r *http.Request) (time.Duration, error) {
	// If timeout is passed as argument, use it's value over default timeout
	if t := r.URL.Query().Get("timeout"); t != "" {
		return time.ParseDuration(t)
	}
	// Fallback on default timeout
	return h.RequestTimeout, nil
}

// getContext creates a context with timeout if timeout is specified in the request or
// server configuration. The context will automatically be canceled as soon as passed
// request connection will be closed.
func (h *Handler) getContext(w http.ResponseWriter, r *http.Request) (context.Context, *Error) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	// Get request timeout request or server config
	timeout, err := h.getTimeout(r)
	if err != nil {
		return nil, &Error{422, fmt.Sprintf("Cannot parse timeout parameter: %s", err), nil}
	}
	if timeout > 0 {
		// Setup a net/context with timeout if time has been specified in either request
		// or server configuration
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
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
	return ctx, nil
}

// ServeHTTP handle requests as a http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Skip body if method is HEAD
	skipBody := r.Method == "HEAD"
	ctx, err := h.getContext(w, r)
	if err != nil {
		h.ResponseSender.SendError(w, err.Code, err, skipBody)
		return
	}
	route, err := h.router.FindRoute(ctx, r)
	if err != nil {
		h.ResponseSender.SendError(w, err.Code, err, skipBody)
		return
	}
	// Store the route and the router in the context
	ctx = contextWithRoute(ctx, route)
	ctx = contextWithRouter(ctx, h.router)
	status, headers, res := processRequest(ctx, route, &request{r})
	for key, values := range headers {
		for _, value := range values {
			w.Header().Set(key, value)
		}
	}
	switch res := res.(type) {
	case *Item:
		h.ResponseSender.SendItem(w, status, res, skipBody)
	case *ItemList:
		h.ResponseSender.SendList(w, status, res, skipBody)
	case *Error:
		h.ResponseSender.SendError(w, status, res, skipBody)
	case error:
		h.ResponseSender.SendError(w, status, res, skipBody)
	default:
		h.ResponseSender.Send(w, status, res)
	}
}
