package rest

import (
	"net/http"
	"strings"

	"golang.org/x/net/context"
)

// processRequest calls the right method on a requestHandler for the given route and return
// either an Item, an ItemList, nil or an Error
func processRequest(ctx context.Context, route RouteMatch, r requestHandler) (status int, headers http.Header, body interface{}) {
	resource := route.Resource()
	if resource == nil {
		return 404, nil, &Error{404, "Resource Not Found", nil}
	}
	if resource.handler == nil {
		return 501, nil, &Error{501, "No handler defined", nil}
	}
	if id := route.ResourceID(); id != nil {
		// Item request
		switch route.Method {
		case "OPTIONS":
			headers = http.Header{}
			methods := []string{}
			if resource.conf.isModeAllowed(Read) {
				methods = append(methods, "HEAD", "GET")
			}
			if resource.conf.isModeAllowed(Create) || resource.conf.isModeAllowed(Replace) {
				methods = append(methods, "PUT")
			}
			if resource.conf.isModeAllowed(Update) {
				methods = append(methods, "PATCH")
				// See http://tools.ietf.org/html/rfc5789#section-3
				headers.Set("Allow-Patch", "application/json")
			}
			if resource.conf.isModeAllowed(Update) {
				methods = append(methods, "DELETE")
			}
			headers.Set("Allow", strings.Join(methods, ", "))
			return 200, headers, nil
		case "HEAD", "GET":
			if !resource.conf.isModeAllowed(Read) {
				return InvalidMethodError.Code, nil, InvalidMethodError
			}
			return r.itemGet(ctx, route)
		case "PUT":
			if !resource.conf.isModeAllowed(Create) && !resource.conf.isModeAllowed(Replace) {
				return InvalidMethodError.Code, nil, InvalidMethodError
			}
			return r.itemPut(ctx, route)
		case "PATCH":
			if !resource.conf.isModeAllowed(Update) {
				return InvalidMethodError.Code, nil, InvalidMethodError
			}
			return r.itemPatch(ctx, route)
		case "DELETE":
			if !resource.conf.isModeAllowed(Delete) {
				return InvalidMethodError.Code, nil, InvalidMethodError
			}
			return r.itemDelete(ctx, route)
		default:
			return InvalidMethodError.Code, nil, InvalidMethodError
		}
	} else {
		// Collection request
		switch route.Method {
		case "OPTIONS":
			headers = http.Header{}
			methods := []string{}
			if resource.conf.isModeAllowed(List) {
				methods = append(methods, "HEAD", "GET")
			}
			if resource.conf.isModeAllowed(Create) {
				methods = append(methods, "POST")
			}
			if resource.conf.isModeAllowed(Clear) {
				methods = append(methods, "DELETE")
			}
			headers.Set("Allow", strings.Join(methods, ", "))
			return 200, headers, nil
		case "HEAD", "GET":
			if !resource.conf.isModeAllowed(List) {
				return InvalidMethodError.Code, nil, InvalidMethodError
			}
			return r.listGet(ctx, route)
		case "POST":
			if !resource.conf.isModeAllowed(Create) {
				return InvalidMethodError.Code, nil, InvalidMethodError
			}
			return r.listPost(ctx, route)
		case "DELETE":
			if !resource.conf.isModeAllowed(Clear) {
				return InvalidMethodError.Code, nil, InvalidMethodError
			}
			return r.listDelete(ctx, route)
		default:
			return InvalidMethodError.Code, nil, InvalidMethodError
		}
	}
}
