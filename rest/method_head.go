package rest

import (
	"context"
	"net/http"

	"github.com/rs/rest-layer/resource"
)

// listHead handles HEAD resquests on a resource URL. Returns no payload
func listHead(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	forceTotal := false
	rsc := route.Resource()
	switch rsc.Conf().ForceTotal {
	case resource.TotalOptIn:
		forceTotal = route.Params.Get("total") == "1"
	case resource.TotalAlways:
		forceTotal = true
	case resource.TotalDenied:
		if route.Params.Get("total") == "1" {
			return 422, nil, &Error{422, "Cannot use `total' parameter: denied by configuration", nil}
		}
	}
	q, e := route.Query()
	if e != nil {
		return e.Code, nil, e
	}
	list := &resource.ItemList{
		Total: -1,
		Items: []*resource.Item{},
	}
	var err error
	if forceTotal {
		list.Total, err = rsc.Count(ctx, q)
		// If Storer doesn't implement Counter interface,
		// fallback to listGet implementation
		if err == resource.ErrNotImplemented {
			return listGet(ctx, r, route)
		}
	}
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	if win := q.Window; win != nil {
		if win.Offset > 0 {
			list.Offset = win.Offset
		}
		if win.Limit >= 0 {
			list.Limit = win.Limit
		}
	}
	return 200, nil, list
}
