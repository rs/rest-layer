package rest

import (
	"net/http"
	"strings"

	"github.com/rs/rest-layer/resource"

	"golang.org/x/net/context"
)

// getItemAllowHeader builds a Allow header based on the resource configuration.
func getItemAllowHeader(conf resource.Conf) http.Header {
	methods := []string{}
	headers := http.Header{}
	// Methods are sorted
	if conf.IsModeAllowed(resource.Update) {
		methods = append(methods, "DELETE")
	}
	if conf.IsModeAllowed(resource.Read) {
		methods = append(methods, "GET, HEAD")
	}
	if conf.IsModeAllowed(resource.Update) {
		methods = append(methods, "PATCH")
		// See http://tools.ietf.org/html/rfc5789#section-3
		headers.Set("Allow-Patch", "application/json")
	}
	if conf.IsModeAllowed(resource.Create) || conf.IsModeAllowed(resource.Replace) {
		methods = append(methods, "PUT")
	}
	if len(methods) > 0 {
		headers.Set("Allow", strings.Join(methods, ", "))
	}
	return headers
}

// itemOptions handles OPTIONS requests on a item URL
func (r *request) itemOptions(ctx context.Context, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	rsrc := route.Resource()
	if rsrc == nil {
		return 404, nil, &Error{404, "Resource Not Found", nil}
	}
	conf := rsrc.Conf()
	headers = getItemAllowHeader(conf)
	return 200, headers, nil
}
