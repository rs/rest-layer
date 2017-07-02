package rest

import (
	"context"
	"net/http"
)

// itemOptions handles OPTIONS requests on a item URL.
func itemOptions(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	rsrc := route.Resource()
	if rsrc == nil {
		return 404, nil, &Error{404, "Resource Not Found", nil}
	}
	conf := rsrc.Conf()
	headers = http.Header{}
	setAllowHeader(headers, true, conf)
	return 200, headers, nil
}
