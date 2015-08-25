package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"golang.org/x/net/context"
)

// Handler is a net/http compatible handler used to serve the configured REST API
type Handler struct {
	// ResponseSender can be changed to extend the DefaultResponseSender
	ResponseSender ResponseSender
	// RequestTimeout is the default timeout for requests after which the whole request
	// is abandonned. The default value is no timeout.
	RequestTimeout time.Duration
	// index stores the resource router
	index resource.Index
	// mw is the list of middlewares attached to this REST handler
	mw []Middleware
}

// NewHandler creates an new REST API HTTP handler with the specified resource index
func NewHandler(i resource.Index) (*Handler, error) {
	if c, ok := i.(schema.Compiler); ok {
		if err := c.Compile(); err != nil {
			return nil, err
		}
	}
	h := &Handler{
		ResponseSender: DefaultResponseSender{},
		index:          i,
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
		h.sendResponse(context.Background(), w, 0, http.Header{}, err, skipBody, nil)
		return
	}
	route, err := FindRoute(h.index, r)
	if err != nil {
		h.sendResponse(ctx, w, 0, http.Header{}, err, skipBody, nil)
		return
	}
	defer route.Release()
	// Store the route and the router in the context
	ctx = contextWithRoute(ctx, route)
	ctx = contextWithIndex(ctx, h.index)

	// Call the middleware + the main route handler
	ctx, status, headers, res := h.callMiddlewares(ctx, r, func(ctx context.Context) (context.Context, int, http.Header, interface{}) {
		// Check route's resource parent(s) exists
		// We perform this check after middlewares, so middleware can prepend route.ResourcePath with
		// some other resources like a user by auth. This will ensure this user resource for instance,
		// 1) exists 2) is contained in all subsequent path resources 3) is set on all newly created
		// resource.
		if err := route.ResourcePath.ParentsExist(ctx); err != nil {
			return ctx, 0, http.Header{}, err
		}
		// Execute the main route handler
		status, headers, body := processRequest(ctx, route, &request{r})
		if headers == nil {
			headers = http.Header{}
		}
		return ctx, status, headers, body
	})
	var v schema.Validator
	if route.Resource() != nil {
		v = route.Resource().Validator()
	}
	h.sendResponse(ctx, w, status, headers, res, skipBody, v)
}

// sendResponse routes the type of response on the right response sender method for
// internally supported types.
func (h *Handler) sendResponse(ctx context.Context, w http.ResponseWriter, status int, headers http.Header, res interface{}, skipBody bool, validator schema.Validator) {
	var body interface{}
	switch res := res.(type) {
	case *resource.Item:
		if s, ok := validator.(schema.Serializer); ok {
			// Prepare the payload for marshaling by calling eventual field serializers
			if err := s.Serialize(res.Payload); err != nil {
				err = fmt.Errorf("Error while preparing item: %s", err.Error())
				h.sendResponse(ctx, w, 0, http.Header{}, err, skipBody, validator)
			}
		}
		ctx, body = h.ResponseSender.SendItem(ctx, headers, res, skipBody)
	case *resource.ItemList:
		if s, ok := validator.(schema.Serializer); ok {
			// Prepare the payload for marshaling by calling eventual field serializers
			for i, item := range res.Items {
				if err := s.Serialize(item.Payload); err != nil {
					err = fmt.Errorf("Error while preparing item #%d: %s", i, err.Error())
					h.sendResponse(ctx, w, 0, http.Header{}, err, skipBody, validator)
				}
			}
		}
		ctx, body = h.ResponseSender.SendList(ctx, headers, res, skipBody)
	case *Error:
		if status == 0 {
			status = res.Code
		}
		ctx, body = h.ResponseSender.SendError(ctx, headers, res, skipBody)
	case error:
		if status == 0 {
			status = 500
		}
		ctx, body = h.ResponseSender.SendError(ctx, headers, res, skipBody)
	default:
		// Let the response sender handle all other types of responses.
		// Even if the default response sender doesn't know how to handle
		// a type, nothing prevents a custom response sender from handling it.
		body = res
	}
	// Send the ResponseWriter
	h.ResponseSender.Send(ctx, w, status, headers, body)
}
