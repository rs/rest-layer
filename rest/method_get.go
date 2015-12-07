package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"

	"golang.org/x/net/context"
)

// listGet handles GET resquests on a resource URL
func (r *request) listGet(ctx context.Context, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	page := 1
	perPage := 0
	if route.Method != "HEAD" {
		if l := route.Resource().Conf().PaginationDefaultLimit; l > 0 {
			perPage = l
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
	lookup, e := route.Lookup()
	if e != nil {
		return e.Code, nil, e
	}
	list, err := route.Resource().Find(ctx, lookup, page, perPage)
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	for _, item := range list.Items {
		item.Payload, err = lookup.ApplySelector(route.Resource(), item.Payload, func(path string, value interface{}) (*resource.Resource, map[string]interface{}, error) {
			router, ok := IndexFromContext(ctx)
			if !ok {
				return nil, nil, errors.New("router not available in context")
			}
			rsrc, _, found := router.GetResource(path)
			if !found {
				return nil, nil, fmt.Errorf("invalid resource reference: %s", path)
			}
			l := resource.NewLookup()
			l.AddQuery(schema.Query{schema.Equal{Field: "id", Value: value}})
			list, _ := rsrc.Find(ctx, l, 1, 1)
			if len(list.Items) == 1 {
				item := list.Items[0]
				return rsrc, item.Payload, nil
			}
			// If no item found, just return an empty dict so we don't error the main request
			return rsrc, map[string]interface{}{}, nil
		})
		if err != nil {
			e = NewError(err)
			return e.Code, nil, e
		}
	}
	return 200, nil, list
}
