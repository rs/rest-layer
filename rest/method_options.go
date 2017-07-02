package rest

import (
	"context"
	"net/http"
)

// listOptions handles OPTIONS requests on a resource URL.
func listOptions(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	rsrc := route.Resource()
	if rsrc == nil {
		return 404, nil, &Error{404, "Resource Not Found", nil}
	}
	conf := rsrc.Conf()
	headers = http.Header{}
	setAllowHeader(headers, false, conf)
	return 200, headers, nil
}
