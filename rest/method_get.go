package rest

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
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
		if l, found, err := getUintParam(route.Params, "limit"); found {
			if err != nil {
				return 422, nil, err
			}
			limit = l
		}
		skip := 0
		if s, found, err := getUintParam(route.Params, "skip"); found {
			if err != nil {
				return 422, nil, err
			}
			skip = s
		}
		page := 1
		if p, found, err := getUintParam(route.Params, "page"); found {
			if err != nil {
				return 422, nil, err
			}
			page = p
		}
		if page > 1 && limit <= 0 {
			return 422, nil, &Error{422, "Cannot use `page' parameter with no `limit' parameter on a resource with no default pagination size", nil}
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

func getUintParam(params url.Values, name string) (int, bool, error) {
	if v := params.Get(name); v != "" {
		i, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, true, &Error{422, fmt.Sprintf("Invalid `%s` parameter", name), nil}
		}
		return int(i), true, nil
	}
	return 0, false, nil
}
