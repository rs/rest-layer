package rest

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/rs/rest-layer/resource"
)

// listGet handles GET resquests on a resource URL.
func listGet(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
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
	var list *resource.ItemList
	var err error
	if forceTotal {
		list, err = rsc.FindWithTotal(ctx, q)
	} else {
		list, err = rsc.Find(ctx, q)
	}
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	if win := q.Window; win != nil && win.Offset > 0 {
		list.Offset = win.Offset
	}
	for _, item := range list.Items {
		item.Payload, err = q.Projection.Eval(ctx, item.Payload, restResource{rsc})
		if err != nil {
			e = NewError(err)
			return e.Code, nil, e
		}
	}
	return 200, nil, list
}

func getUintParam(params url.Values, name string) (int, bool, error) {
	if v := params.Get(name); v != "" {
		i, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, true, errors.New("must be positive integer")
		}
		return int(i), true, nil
	}
	return 0, false, nil
}
