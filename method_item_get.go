package rest

import (
	"net/http"
	"time"

	"golang.org/x/net/context"
)

// itemGet handles GET and HEAD resquests on an item URL
func (r *request) itemGet(ctx context.Context, route RouteMatch) (status int, headers http.Header, body interface{}) {
	lookup, err := route.Lookup()
	if err != nil {
		return err.Code, nil, err
	}
	l, err := route.Resource().handler.Find(ctx, lookup, 1, 1)
	if err != nil {
		return err.Code, nil, err
	} else if len(l.Items) == 0 {
		return NotFoundError.Code, nil, NotFoundError
	}
	item := l.Items[0]
	// Handle conditional request: If-None-Match
	if r.req.Header.Get("If-None-Match") == item.Etag {
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
	return 200, nil, item
}
