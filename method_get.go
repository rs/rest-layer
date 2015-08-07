package rest

import (
	"net/http"
	"strconv"

	"golang.org/x/net/context"
)

// listGet handles GET resquests on a resource URL
func (r *request) listGet(ctx context.Context, route RouteMatch) (status int, headers http.Header, body interface{}) {
	page := 1
	perPage := 0
	if route.Method != "HEAD" {
		if route.Resource().conf.PaginationDefaultLimit > 0 {
			perPage = route.Resource().conf.PaginationDefaultLimit
		} else {
			// Default value on non HEAD request for perPage is -1 (pagination disabled)
			perPage = -1
		}
		if p := r.req.URL.Query().Get("page"); p != "" {
			i, err := strconv.ParseUint(p, 10, 32)
			if err != nil {
				return 422, nil, &Error{422, "Invalid `page` paramter", nil}
			}
			page = int(i)
		}
		if l := r.req.URL.Query().Get("limit"); l != "" {
			i, err := strconv.ParseUint(l, 10, 32)
			if err != nil {
				return 422, nil, &Error{422, "Invalid `limit` paramter", nil}
			}
			perPage = int(i)
		}
		if perPage == -1 && page != 1 {
			return 422, nil, &Error{422, "Cannot use `page' parameter with no `limit' paramter on a resource with no default pagination size", nil}
		}
	}
	lookup, err := route.Lookup()
	if err != nil {
		return err.Code, nil, err
	}
	list, err := route.Resource().handler.Find(ctx, lookup, page, perPage)
	if err != nil {
		return err.Code, nil, err
	}
	return 200, nil, list
}
