package rest

import (
	"net/http"
	"strconv"

	"golang.org/x/net/context"
)

// listDelete handles DELETE resquests on a resource URL
func (r *request) listDelete(ctx context.Context, route RouteMatch) (status int, headers http.Header, body interface{}) {
	lookup, err := route.Lookup()
	if err != nil {
		return err.Code, nil, err
	}
	total, err := route.Resource().handler.Clear(ctx, lookup)
	if err != nil {
		return err.Code, nil, err
	}
	headers = http.Header{}
	headers.Set("X-Total", strconv.FormatInt(int64(total), 10))
	return 204, headers, nil
}
