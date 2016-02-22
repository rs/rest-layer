package rest

import (
	"net/http"

	"github.com/rs/rest-layer/resource"
	"golang.org/x/net/context"
)

// processRequest calls the right method on a requestHandler for the given route and return
// either an Item, an ItemList, nil or an Error
func processRequest(ctx context.Context, route *RouteMatch, r requestHandler) (status int, headers http.Header, body interface{}) {
	if route.Method == "OPTIONS" {
		if id := route.ResourceID(); id != nil {
			return r.itemOptions(ctx, route)
		}
		return r.listOptions(ctx, route)
	}
	rsrc := route.Resource()
	if rsrc == nil {
		return 404, nil, &Error{404, "Resource Not Found", nil}
	}
	conf := rsrc.Conf()
	if id := route.ResourceID(); id != nil {
		// Item request
		switch route.Method {
		case "HEAD", "GET":
			if !conf.IsModeAllowed(resource.Read) {
				return ErrInvalidMethod.Code, getItemAllowHeader(conf), ErrInvalidMethod
			}
			return r.itemGet(ctx, route)
		case "PUT":
			if !conf.IsModeAllowed(resource.Create) && !conf.IsModeAllowed(resource.Replace) {
				return ErrInvalidMethod.Code, getItemAllowHeader(conf), ErrInvalidMethod
			}
			return r.itemPut(ctx, route)
		case "PATCH":
			if !conf.IsModeAllowed(resource.Update) {
				return ErrInvalidMethod.Code, getItemAllowHeader(conf), ErrInvalidMethod
			}
			return r.itemPatch(ctx, route)
		case "DELETE":
			if !conf.IsModeAllowed(resource.Delete) {
				return ErrInvalidMethod.Code, getItemAllowHeader(conf), ErrInvalidMethod
			}
			return r.itemDelete(ctx, route)
		default:
			return ErrInvalidMethod.Code, getItemAllowHeader(conf), ErrInvalidMethod
		}
	} else {
		// Collection request
		switch route.Method {
		case "HEAD", "GET":
			if !conf.IsModeAllowed(resource.List) {
				return ErrInvalidMethod.Code, getListAllowHeader(conf), ErrInvalidMethod
			}
			return r.listGet(ctx, route)
		case "POST":
			if !conf.IsModeAllowed(resource.Create) {
				return ErrInvalidMethod.Code, getListAllowHeader(conf), ErrInvalidMethod
			}
			return r.listPost(ctx, route)
		case "DELETE":
			if !conf.IsModeAllowed(resource.Clear) {
				return ErrInvalidMethod.Code, getListAllowHeader(conf), ErrInvalidMethod
			}
			return r.listDelete(ctx, route)
		default:
			return ErrInvalidMethod.Code, getListAllowHeader(conf), ErrInvalidMethod
		}
	}
}
