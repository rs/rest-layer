package rest

import (
	"strings"

	"golang.org/x/net/context"
)

// serveRequest calls the right method on a requestHandler for the given route
func serveRequest(ctx context.Context, route route, r requestHandler) {
	if route.resource.handler == nil {
		r.sendError(&Error{501, "No handler defined", nil})
		return
	}
	if _, found := route.fields["id"]; found {
		// Item request
		switch route.method {
		case "OPTIONS":
			methods := []string{}
			if route.resource.conf.isModeAllowed(Read) {
				methods = append(methods, "HEAD", "GET")
			}
			if route.resource.conf.isModeAllowed(Create) || route.resource.conf.isModeAllowed(Replace) {
				methods = append(methods, "PUT")
			}
			if route.resource.conf.isModeAllowed(Update) {
				methods = append(methods, "PATCH")
				// See http://tools.ietf.org/html/rfc5789#section-3
				r.setHeader("Allow-Patch", "application/json")
			}
			if route.resource.conf.isModeAllowed(Update) {
				methods = append(methods, "DELETE")
			}
			r.setHeader("Allow", strings.Join(methods, ", "))
		case "HEAD", "GET":
			if !route.resource.conf.isModeAllowed(Read) {
				r.sendError(InvalidMethodError)
				return
			}
			r.itemGet(ctx, route)
		case "PUT":
			if !route.resource.conf.isModeAllowed(Create) && !route.resource.conf.isModeAllowed(Replace) {
				r.sendError(InvalidMethodError)
				return
			}
			r.itemPut(ctx, route)
		case "PATCH":
			if !route.resource.conf.isModeAllowed(Update) {
				r.sendError(InvalidMethodError)
				return
			}
			r.itemPatch(ctx, route)
		case "DELETE":
			if !route.resource.conf.isModeAllowed(Delete) {
				r.sendError(InvalidMethodError)
				return
			}
			r.itemDelete(ctx, route)
		default:
			r.sendError(InvalidMethodError)
		}
	} else {
		// Collection request
		switch route.method {
		case "OPTIONS":
			methods := []string{}
			if route.resource.conf.isModeAllowed(List) {
				methods = append(methods, "HEAD", "GET")
			}
			if route.resource.conf.isModeAllowed(Create) {
				methods = append(methods, "POST")
			}
			if route.resource.conf.isModeAllowed(Clear) {
				methods = append(methods, "DELETE")
			}
			r.setHeader("Allow", strings.Join(methods, ", "))
		case "HEAD", "GET":
			if !route.resource.conf.isModeAllowed(List) {
				r.sendError(InvalidMethodError)
				return
			}
			r.listGet(ctx, route)
		case "POST":
			if !route.resource.conf.isModeAllowed(Create) {
				r.sendError(InvalidMethodError)
				return
			}
			r.listPost(ctx, route)
		case "DELETE":
			if !route.resource.conf.isModeAllowed(Clear) {
				r.sendError(InvalidMethodError)
				return
			}
			r.listDelete(ctx, route)
		default:
			r.sendError(InvalidMethodError)
		}
	}
}
