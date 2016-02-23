package rest

import (
	"net/http"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"golang.org/x/net/context"
)

// Handler is a net/http compatible handler used to serve the configured REST API
type Handler struct {
	// ResponseFormatter can be changed to extend the DefaultResponseFormatter
	ResponseFormatter ResponseFormatter
	// ResponseSender can be changed to extend the DefaultResponseSender
	ResponseSender ResponseSender
	// index stores the resource router
	index resource.Index
	// mw is the list of middlewares attached to this REST handler
	mw []Middleware
}

type methodHandler func(ctx context.Context, r *http.Request, route *RouteMatch) (int, http.Header, interface{})

// NewHandler creates an new REST API HTTP handler with the specified resource index
func NewHandler(i resource.Index) (*Handler, error) {
	if c, ok := i.(schema.Compiler); ok {
		if err := c.Compile(); err != nil {
			return nil, err
		}
	}
	h := &Handler{
		ResponseFormatter: DefaultResponseFormatter{},
		ResponseSender:    DefaultResponseSender{},
		index:             i,
	}
	return h, nil
}

// getContext creates a context for the request to add net/context support when used as a
// standard http.Handler, without net/context support. The context will automatically be
// canceled as soon as passed request connection will be closed.
func (h *Handler) getContext(w http.ResponseWriter, r *http.Request) (context.Context, *Error) {
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
	return ctx, nil
}

// ServeHTTP handles requests as a http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Skip body if method is HEAD
	skipBody := r.Method == "HEAD"
	ctx, err := h.getContext(w, r)
	if err != nil {
		h.sendResponse(context.Background(), w, 0, http.Header{}, err, skipBody, nil)
		return
	}
	h.ServeHTTPC(ctx, w, r)
}

// ServeHTTPC handles requests as a xhandler.HandlerC
func (h *Handler) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// Skip body if method is HEAD
	skipBody := r.Method == "HEAD"
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
		// Execute the main route handler
		status, headers, body := routeHandler(ctx, r, route)
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

// routeHandler executes the appropriate method handler for the request if allowed by the route configuration
func routeHandler(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	// Check route's resource parent(s) exists
	// We perform this check after middlewares, so middleware can prepend route.ResourcePath with
	// some other resources like a user by auth. This will ensure this user resource for instance,
	// 1) exists 2) is contained in all subsequent path resources 3) is set on all newly created
	// resource.
	if err := route.ResourcePath.ParentsExist(ctx); err != nil {
		return 0, http.Header{}, err
	}
	rsrc := route.Resource()
	if rsrc == nil {
		return http.StatusNotFound, nil, &Error{http.StatusNotFound, "Resource Not Found", nil}
	}
	conf := rsrc.Conf()
	isItem := route.ResourceID() != nil
	mh := getAllowedMethodHandler(isItem, route.Method, conf)
	if mh == nil {
		headers = http.Header{}
		setAllowHeader(headers, isItem, conf)
		return ErrInvalidMethod.Code, headers, ErrInvalidMethod
	}
	return mh(ctx, r, route)
}

// sendResponse format and send the API response
func (h *Handler) sendResponse(ctx context.Context, w http.ResponseWriter, status int, headers http.Header, res interface{}, skipBody bool, validator schema.Validator) {
	ctx, status, body := formatResponse(ctx, h.ResponseFormatter, w, status, headers, res, skipBody, validator)
	h.ResponseSender.Send(ctx, w, status, headers, body)
}
