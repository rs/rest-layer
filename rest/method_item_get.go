package rest

import (
	"net/http"
	"time"

	"golang.org/x/net/context"
)

// itemGet handles GET and HEAD resquests on an item URL
func (r *request) itemGet(ctx context.Context, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	lookup, e := route.Lookup()
	if e != nil {
		return e.Code, nil, e
	}
	l, err := route.Resource().Find(ctx, lookup, 1, 1)
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	} else if len(l.Items) == 0 {
		return ErrNotFound.Code, nil, ErrNotFound
	}
	item := l.Items[0]
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
	item.Payload = lookup.ApplySelector(item.Payload)
	return 200, nil, item
}
