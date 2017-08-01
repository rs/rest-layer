package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/rest-layer/schema/query"
)

// itemGet handles GET and HEAD resquests on an item URL.
func itemGet(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	q, e := route.Query()
	if e != nil {
		return e.Code, nil, e
	}
	rsrc := route.Resource()
	q.Window = &query.Window{Limit: 1}
	list, err := rsrc.Find(ctx, q)
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	} else if len(list.Items) == 0 {
		return ErrNotFound.Code, nil, ErrNotFound
	}
	item := list.Items[0]
	// Handle conditional request: If-None-Match.
	if compareEtag(r.Header.Get("If-None-Match"), item.ETag) {
		return 304, nil, nil
	}
	// Handle conditional request: If-Modified-Since.
	if r.Header.Get("If-Modified-Since") != "" {
		if ifModTime, err := time.Parse(time.RFC1123, r.Header.Get("If-Modified-Since")); err != nil {
			return 400, nil, &Error{400, "Invalid If-Modified-Since header", nil}
		} else if u := item.Updated.Truncate(time.Second); u.Equal(ifModTime) || u.Before(ifModTime) {
			// Item's update time is truncated to the second because RFC1123
			// doesn't support more.
			return 304, nil, nil
		}
	}
	item.Payload, err = q.Projection.Eval(ctx, item.Payload, restResource{rsrc})
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	return 200, nil, item
}
