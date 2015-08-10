package rest

import (
	"net/http"
	"strings"

	"github.com/rs/rest-layer/resource"
	"golang.org/x/net/context"
)

// processRequest calls the right method on a requestHandler for the given route and return
// either an Item, an ItemList, nil or an Error
func processRequest(ctx context.Context, route RouteMatch, r requestHandler) (status int, headers http.Header, body interface{}) {
	rsrc := route.Resource()
	if rsrc == nil {
		return 404, nil, &Error{404, "Resource Not Found", nil}
	}
	conf := rsrc.Conf()
	if id := route.ResourceID(); id != nil {
		// Item request
		switch route.Method {
		case "OPTIONS":
			headers = http.Header{}
			methods := []string{}
			if conf.IsModeAllowed(resource.Read) {
				methods = append(methods, "HEAD", "GET")
			}
			if conf.IsModeAllowed(resource.Create) || conf.IsModeAllowed(resource.Replace) {
				methods = append(methods, "PUT")
			}
			if conf.IsModeAllowed(resource.Update) {
				methods = append(methods, "PATCH")
				// See http://tools.ietf.org/html/rfc5789#section-3
				headers.Set("Allow-Patch", "application/json")
			}
			if conf.IsModeAllowed(resource.Update) {
				methods = append(methods, "DELETE")
			}
			headers.Set("Allow", strings.Join(methods, ", "))
			return 200, headers, nil
		case "HEAD", "GET":
			if !conf.IsModeAllowed(resource.Read) {
				return ErrInvalidMethod.Code, nil, ErrInvalidMethod
			}
			return r.itemGet(ctx, route)
		case "PUT":
			if !conf.IsModeAllowed(resource.Create) && !conf.IsModeAllowed(resource.Replace) {
				return ErrInvalidMethod.Code, nil, ErrInvalidMethod
			}
			return r.itemPut(ctx, route)
		case "PATCH":
			if !conf.IsModeAllowed(resource.Update) {
				return ErrInvalidMethod.Code, nil, ErrInvalidMethod
			}
			return r.itemPatch(ctx, route)
		case "DELETE":
			if !conf.IsModeAllowed(resource.Delete) {
				return ErrInvalidMethod.Code, nil, ErrInvalidMethod
			}
			return r.itemDelete(ctx, route)
		default:
			return ErrInvalidMethod.Code, nil, ErrInvalidMethod
		}
	} else {
		// Collection request
		switch route.Method {
		case "OPTIONS":
			headers = http.Header{}
			methods := []string{}
			if conf.IsModeAllowed(resource.List) {
				methods = append(methods, "HEAD", "GET")
			}
			if conf.IsModeAllowed(resource.Create) {
				methods = append(methods, "POST")
			}
			if conf.IsModeAllowed(resource.Clear) {
				methods = append(methods, "DELETE")
			}
			headers.Set("Allow", strings.Join(methods, ", "))
			return 200, headers, nil
		case "HEAD", "GET":
			if !conf.IsModeAllowed(resource.List) {
				return ErrInvalidMethod.Code, nil, ErrInvalidMethod
			}
			return r.listGet(ctx, route)
		case "POST":
			if !conf.IsModeAllowed(resource.Create) {
				return ErrInvalidMethod.Code, nil, ErrInvalidMethod
			}
			return r.listPost(ctx, route)
		case "DELETE":
			if !conf.IsModeAllowed(resource.Clear) {
				return ErrInvalidMethod.Code, nil, ErrInvalidMethod
			}
			return r.listDelete(ctx, route)
		default:
			return ErrInvalidMethod.Code, nil, ErrInvalidMethod
		}
	}
}
