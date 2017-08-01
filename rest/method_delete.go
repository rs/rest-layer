package rest

import (
	"context"
	"net/http"
	"strconv"
)

// listDelete handles DELETE resquests on a resource URL.
func listDelete(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	q, e := route.Query()
	if e != nil {
		return e.Code, nil, e
	}
	total, err := route.Resource().Clear(ctx, q)
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	headers = http.Header{}
	headers.Set("X-Total", strconv.Itoa(total))
	return 204, headers, nil
}
