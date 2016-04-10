package rest

import (
	"net/http"
	"strconv"

	"golang.org/x/net/context"
)

// listGet handles GET resquests on a resource URL
func listGet(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	page := 1
	perPage := 0
	rsrc := route.Resource()
	if route.Method != "HEAD" {
		if l := rsrc.Conf().PaginationDefaultLimit; l > 0 {
			perPage = l
		} else {
			// Default value on non HEAD request for perPage is -1 (pagination disabled)
			perPage = -1
		}
		if p := route.Params.Get("page"); p != "" {
			i, err := strconv.ParseUint(p, 10, 32)
			if err != nil {
				return 422, nil, &Error{422, "Invalid `page` parameter", nil}
			}
			page = int(i)
		}
		if l := route.Params.Get("limit"); l != "" {
			i, err := strconv.ParseUint(l, 10, 32)
			if err != nil {
				return 422, nil, &Error{422, "Invalid `limit` parameter", nil}
			}
			perPage = int(i)
		}
		if perPage == -1 && page != 1 {
			return 422, nil, &Error{422, "Cannot use `page' parameter with no `limit' parameter on a resource with no default pagination size", nil}
		}
	}
	lookup, e := route.Lookup()
	if e != nil {
		return e.Code, nil, e
	}
	list, err := rsrc.Find(ctx, lookup, page, perPage)
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	for _, item := range list.Items {
		item.Payload, err = lookup.ApplySelector(ctx, rsrc, item.Payload, getReferenceResolver(ctx, rsrc))
		if err != nil {
			e = NewError(err)
			return e.Code, nil, e
		}
	}
	return 200, nil, list
}
