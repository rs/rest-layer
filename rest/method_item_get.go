package rest

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"

	"golang.org/x/net/context"
)

// itemGet handles GET and HEAD resquests on an item URL
func (r *request) itemGet(ctx context.Context, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	lookup, e := route.Lookup()
	if e != nil {
		return e.Code, nil, e
	}
	list, err := route.Resource().Find(ctx, lookup, 1, 1)
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	} else if len(list.Items) == 0 {
		return ErrNotFound.Code, nil, ErrNotFound
	}
	item := list.Items[0]
	// Handle conditional request: If-None-Match
	if r.req.Header.Get("If-None-Match") == item.ETag {
		return 304, nil, nil
	}
	// Handle conditional request: If-Modified-Since
	if r.req.Header.Get("If-Modified-Since") != "" {
		if ifModTime, err := time.Parse(time.RFC1123, r.req.Header.Get("If-Modified-Since")); err != nil {
			return 400, nil, &Error{400, "Invalid If-Modified-Since header", nil}
		} else if item.Updated.Equal(ifModTime) || item.Updated.Before(ifModTime) {
			return 304, nil, nil
		}
	}
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
	return 200, nil, item
}
