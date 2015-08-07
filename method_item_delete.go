package rest

import (
	"net/http"

	"golang.org/x/net/context"
)

// itemDelete handles DELETE resquests on an item URL
func (r *request) itemDelete(ctx context.Context, route RouteMatch) (status int, headers http.Header, body interface{}) {
	lookup, err := route.Lookup()
	if err != nil {
		return err.Code, nil, err
	}
	l, err := route.Resource().handler.Find(ctx, lookup, 1, 1)
	if err != nil {
		return err.Code, nil, err
	}
	if len(l.Items) == 0 {
		return NotFoundError.Code, nil, NotFoundError
	}
	original := l.Items[0]
	// If-Match / If-Unmodified-Since handling
	if err := r.checkIntegrityRequest(original); err != nil {
		return err.Code, nil, err
	}
	if err := route.Resource().handler.Delete(ctx, original); err != nil {
		return err.Code, nil, err
	}
	return 204, nil, nil
}
