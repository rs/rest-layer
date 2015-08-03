package rest

import (
	"strconv"

	"golang.org/x/net/context"
)

// listGet handles GET resquests on a resource URL
func (r *request) listGet(ctx context.Context, route route) {
	page := 1
	perPage := 0
	if !r.skipBody {
		if route.resource.conf.PaginationDefaultLimit > 0 {
			perPage = route.resource.conf.PaginationDefaultLimit
		} else {
			// Default value on non HEAD request for perPage is -1 (pagination disabled)
			perPage = -1
		}
		if p := r.req.URL.Query().Get("page"); p != "" {
			i, err := strconv.ParseUint(p, 10, 32)
			if err != nil {
				r.sendError(&Error{422, "Invalid `page` paramter", nil})
				return
			}
			page = int(i)
		}
		if l := r.req.URL.Query().Get("limit"); l != "" {
			i, err := strconv.ParseUint(l, 10, 32)
			if err != nil {
				r.sendError(&Error{422, "Invalid `limit` paramter", nil})
				return
			}
			perPage = int(i)
		}
		if perPage == -1 && page != 1 {
			r.sendError(&Error{422, "Cannot use `page' parameter with no `limit' paramter on a resource with no default pagination size", nil})
		}
	}
	lookup, err := route.lookup()
	if err != nil {
		r.sendError(err)
		return
	}
	list, err := route.resource.handler.Find(lookup, page, perPage, ctx)
	if err != nil {
		r.sendError(err)
		return
	}
	r.sendList(list)
}
