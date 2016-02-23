package rest

import (
	"net/http"
	"strings"

	"github.com/rs/rest-layer/resource"

	"golang.org/x/net/context"
)

// getListAllowHeader builds a Allow header based on the resource configuration.
func getListAllowHeader(conf resource.Conf) http.Header {
	methods := []string{}
	headers := http.Header{}
	// Methods are sorted
	if conf.IsModeAllowed(resource.Clear) {
		methods = append(methods, "DELETE")
	}
	if conf.IsModeAllowed(resource.List) {
		methods = append(methods, "GET, HEAD")
	}
	if conf.IsModeAllowed(resource.Create) {
		methods = append(methods, "POST")
	}
	if len(methods) > 0 {
		headers.Set("Allow", strings.Join(methods, ", "))
	}
	return headers
}

// listOptions handles OPTIONS requests on a resource URL
func listOptions(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	rsrc := route.Resource()
	if rsrc == nil {
		return 404, nil, &Error{404, "Resource Not Found", nil}
	}
	conf := rsrc.Conf()
	headers = getListAllowHeader(conf)
	return 200, headers, nil
}
