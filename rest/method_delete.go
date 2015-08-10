package rest

import (
	"net/http"
	"strconv"

	"golang.org/x/net/context"
)

// listDelete handles DELETE resquests on a resource URL
func (r *request) listDelete(ctx context.Context, route RouteMatch) (status int, headers http.Header, body interface{}) {
	lookup, e := route.Lookup()
	if e != nil {
		return e.Code, nil, e
	}
	total, err := route.Resource().Clear(ctx, lookup)
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	headers = http.Header{}
	headers.Set("X-Total", strconv.FormatInt(int64(total), 10))
	return 204, headers, nil
}
