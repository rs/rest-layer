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
	page := -1
	isSetOffset := false
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
		if o := route.Params.Get("offset"); o != "" {
			i, err := strconv.ParseUint(o, 10, 32)
			if err != nil {
				return 422, nil, &Error{422, "Invalid `offset` parameter", nil}
			}
			offset = int(i)
			isSetOffset = true
		}
		if p := route.Params.Get("page"); p != "" {
			if isSetOffset {
				return 422, nil, &Error{422, "Cannot use `page' parameter together with `offset` parameter.", nil}
			}
			i, err := strconv.ParseUint(p, 10, 32)
			if err != nil {
				return 422, nil, &Error{422, "Invalid `page` parameter", nil}
			}
			page = int(i)
			if limit <= 0 {
				return 422, nil, &Error{422, "Cannot use `page' parameter with no `limit' parameter on a resource with no default pagination size", nil}
			}
			offset = page * limit
		}
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
	// Now we can set Total.
	list.Total = len(list.Items)

	// Set appropriate fields.
	if page >= 0 {
		list.Page = page
		list.Skip = (page * limit) - limit
	} else {
		if isSetOffset {
			list.Skip = offset
		} else {
			list.Skip = 0
		}
	}
	for _, item := range list.Items {
		item.Payload, err = lookup.ApplySelector(ctx, rsrc.Validator(), item.Payload, getReferenceResolver(ctx, rsrc))
		if err != nil {
			e = NewError(err)
			return e.Code, nil, e
		}
	}
	return 200, nil, list
}
