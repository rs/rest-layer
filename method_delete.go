package rest

import (
	"strconv"

	"golang.org/x/net/context"
)

// listDelete handles DELETE resquests on a resource URL
func (r *request) listDelete(ctx context.Context, route route) {
	lookup, err := route.lookup()
	if err != nil {
		r.sendError(err)
	}
	if total, err := route.resource.handler.Clear(ctx, lookup); err != nil {
		r.sendError(err)
	} else {
		r.res.Header().Set("X-Total", strconv.FormatInt(int64(total), 10))
		r.send(204, map[string]interface{}{})
	}
}
