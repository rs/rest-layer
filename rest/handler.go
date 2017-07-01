package rest

import (
	"context"
	"net/http"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
)

// Handler is a net/http compatible handler used to serve the configured REST
// API.
type Handler struct {
	// ResponseFormatter can be changed to extend the DefaultResponseFormatter.
	ResponseFormatter ResponseFormatter
	// ResponseSender can be changed to extend the DefaultResponseSender.
	ResponseSender ResponseSender
	// FallbackHandlerFunc is called when REST layer doesn't find a route for
	// the request. If not set, a 404 or 405 standard REST error is returned.
	FallbackHandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request)
	// index stores the resource router.
	index resource.Index
}

type methodHandler func(ctx context.Context, r *http.Request, route *RouteMatch) (int, http.Header, interface{})

// NewHandler creates an new REST API HTTP handler with the specified resource
// index.
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

// ServeHTTP handles requests as a http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.ServeHTTPC(ctx, w, r)
}

// ServeHTTPC handles requests as a xhandler.HandlerC (deprecated).
func (h *Handler) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// Skip body if method is HEAD
	skipBody := r.Method == "HEAD"
	route, err := FindRoute(h.index, r)
	if err != nil {
		if h.FallbackHandlerFunc != nil {
			h.FallbackHandlerFunc(ctx, w, r)
		} else {
			h.sendResponse(ctx, w, 0, http.Header{}, err, skipBody)
		}
		return
	}
	defer route.Release()
	// Store the route and the router in the context
	ctx = contextWithRoute(ctx, route)
	ctx = contextWithIndex(ctx, h.index)

	// Execute the main route handler
	status, headers, body := routeHandler(ctx, r, route)
	if headers == nil {
		headers = http.Header{}
	}
	if h.FallbackHandlerFunc != nil && (body == errResourceNotFound || body == ErrInvalidMethod) {
		h.FallbackHandlerFunc(ctx, w, r)
		return
	}
	h.sendResponse(ctx, w, status, headers, body, skipBody)
}

// routeHandler executes the appropriate method handler for the request if
// allowed by the route configuration.
func routeHandler(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	// Check route's resource parent(s) exists.
	if err := route.ResourcePath.ParentsExist(ctx); err != nil {
		return 0, http.Header{}, err
	}
	rsrc := route.Resource()
	if rsrc == nil {
		return http.StatusNotFound, nil, errResourceNotFound
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

// sendResponse format and send the API response.
func (h *Handler) sendResponse(ctx context.Context, w http.ResponseWriter, status int, headers http.Header, res interface{}, skipBody bool) {
	ctx, status, body := formatResponse(ctx, h.ResponseFormatter, w, status, headers, res, skipBody)
	h.ResponseSender.Send(ctx, w, status, headers, body)
}
