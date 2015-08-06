package rest

import (
	"time"

	"golang.org/x/net/context"
)

// itemGet handles GET and HEAD resquests on an item URL
func (r *request) itemGet(ctx context.Context, route route) {
	lookup, err := route.lookup()
	if err != nil {
		r.sendError(err)
	}
	l, err := route.resource.handler.Find(ctx, lookup, 1, 1)
	if err != nil {
		r.sendError(err)
		return
	} else if len(l.Items) == 0 {
		r.sendError(NotFoundError)
		return
	}
	item := l.Items[0]
	// Handle conditional request: If-None-Match
	if r.req.Header.Get("If-None-Match") == item.Etag {
		r.send(304, nil)
		return
	}
	// Handle conditional request: If-Modified-Since
	if r.req.Header.Get("If-Modified-Since") != "" {
		if ifModTime, err := time.Parse(time.RFC1123, r.req.Header.Get("If-Modified-Since")); err != nil {
			r.sendError(&Error{400, "Invalid If-Modified-Since header", nil})
			return
		} else if item.Updated.Equal(ifModTime) || item.Updated.Before(ifModTime) {
			r.send(304, nil)
			return
		}
	}
	r.sendItem(200, item)
}
