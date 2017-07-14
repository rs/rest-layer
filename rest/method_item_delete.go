package rest

import (
	"context"
	"net/http"

	"github.com/rs/rest-layer/schema/query"
)

// itemDelete handles DELETE resquests on an item URL.
func itemDelete(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	q, e := route.Query()
	if e != nil {
		return e.Code, nil, e
	}
	q.Window = &query.Window{Limit: 1}
	l, err := route.Resource().Find(ctx, q)
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	if len(l.Items) == 0 {
		return ErrNotFound.Code, nil, ErrNotFound
	}
	original := l.Items[0]
	// If-Match / If-Unmodified-Since handling.
	if err := checkIntegrityRequest(r, original); err != nil {
		return err.Code, nil, err
	}
	if err := route.Resource().Delete(ctx, original); err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	return 204, nil, nil
}
