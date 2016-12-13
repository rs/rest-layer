package rest

import (
	"context"
	"net/http"
	"strconv"
)

// listGet handles GET resquests on a resource URL
func listGet(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	offset := 0
	limit := 0
	rsrc := route.Resource()
	if route.Method != "HEAD" {
		if l := rsrc.Conf().PaginationDefaultLimit; l > 0 {
			limit = l
		} else {
			// Default value on non HEAD request for limit is -1 (pagination disabled)
			limit = -1
		}
		if l := route.Params.Get("limit"); l != "" {
			i, err := strconv.ParseUint(l, 10, 32)
			if err != nil {
				return 422, nil, &Error{422, "Invalid `limit` parameter", nil}
			}
			limit = int(i)
		}
		skip := 0
		if o := route.Params.Get("skip"); o != "" {
			i, err := strconv.ParseUint(o, 10, 32)
			if err != nil {
				return 422, nil, &Error{422, "Invalid `skip` parameter", nil}
			}
			skip = int(i)
		}
		page := 1
		if p := route.Params.Get("page"); p != "" {
			i, err := strconv.ParseUint(p, 10, 32)
			if err != nil {
				return 422, nil, &Error{422, "Invalid `page` parameter", nil}
			}
			page = int(i)
			if limit <= 0 {
				return 422, nil, &Error{422, "Cannot use `page' parameter with no `limit' parameter on a resource with no default pagination size", nil}
			}
		}
		offset = (page-1)*limit + skip
	}
	lookup, e := route.Lookup()
	if e != nil {
		return e.Code, nil, e
	}
	list, err := rsrc.Find(ctx, lookup, offset, limit)
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	list.Offset = offset
	for _, item := range list.Items {
		item.Payload, err = lookup.ApplySelector(ctx, rsrc.Validator(), item.Payload, getReferenceResolver(ctx, rsrc))
		if err != nil {
			e = NewError(err)
			return e.Code, nil, e
		}
	}
	return 200, nil, list
}
