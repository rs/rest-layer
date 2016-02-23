package rest

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/rest-layer/resource"

	"golang.org/x/net/context"
)

// itemGet handles GET and HEAD resquests on an item URL
func itemGet(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
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
	if compareEtag(r.Header.Get("If-None-Match"), item.ETag) {
		return 304, nil, nil
	}
	// Handle conditional request: If-Modified-Since
	if r.Header.Get("If-Modified-Since") != "" {
		if ifModTime, err := time.Parse(time.RFC1123, r.Header.Get("If-Modified-Since")); err != nil {
			return 400, nil, &Error{400, "Invalid If-Modified-Since header", nil}
		} else if item.Updated.Equal(ifModTime) || item.Updated.Before(ifModTime) {
			return 304, nil, nil
		}
	}
	item.Payload, err = lookup.ApplySelector(ctx, route.Resource(), item.Payload, func(path string) (*resource.Resource, error) {
		router, ok := IndexFromContext(ctx)
		if !ok {
			return nil, errors.New("router not available in context")
		}
		rsrc, _, found := router.GetResource(path, route.Resource())
		if !found {
			return nil, fmt.Errorf("invalid resource reference: %s", path)
		}
		return rsrc, err
	})
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	return 200, nil, item
}
