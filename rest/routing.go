package rest

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema/query"
)

// RouteMatch represent a REST request's matched resource with the method to
// apply and its parameters.
type RouteMatch struct {
	// Method is the HTTP method used on the resource.
	Method string
	// ResourcePath is the list of intermediate resources followed by the
	// targeted resource. Each intermediate resource much match all the previous
	// resource components of this path and newly created resources will have
	// their corresponding fields filled with resource path information
	// (resource.field => resource.value).
	ResourcePath ResourcePath
	// Params is the list of client provided parameters (thru query-string or alias).
	Params url.Values
}

type key int

const (
	routeKey key = iota
	indexKey
)

var routePool = sync.Pool{
	New: func() interface{} {
		return &RouteMatch{
			ResourcePath: make(ResourcePath, 0, 2),
		}
	},
}

var errResourceNotFound = &Error{http.StatusNotFound, "Resource Not Found", nil}

func contextWithRoute(ctx context.Context, route *RouteMatch) context.Context {
	return context.WithValue(ctx, routeKey, route)
}

func contextWithIndex(ctx context.Context, index resource.Index) context.Context {
	return context.WithValue(ctx, indexKey, index)
}

// RouteFromContext extracts the matched route from the given net/context.
func RouteFromContext(ctx context.Context) (*RouteMatch, bool) {
	route, ok := ctx.Value(routeKey).(*RouteMatch)
	return route, ok
}

// IndexFromContext extracts the router from the given net/context.
func IndexFromContext(ctx context.Context) (resource.Index, bool) {
	index, ok := ctx.Value(indexKey).(resource.Index)
	return index, ok
}

// FindRoute returns the REST route for the given request.
func FindRoute(index resource.Index, req *http.Request) (*RouteMatch, error) {
	route := routePool.Get().(*RouteMatch)
	route.Method = req.Method
	route.Params = req.URL.Query()

	err := findRoute(req.URL.Path, index, route)
	if err != nil {
		route.Release()
		route = nil
	}
	return route, err
}

// findRoute recursively route a (sub)resource request.
func findRoute(path string, index resource.Index, route *RouteMatch) error {
	// Extract the first component of the path.
	var name string
	name, path = nextPathComponent(path)

	resourcePath := name
	if prefix := route.ResourcePath.Path(); prefix != "" {
		resourcePath = prefix + "." + name
	}

	if rsrc, found := index.GetResource(resourcePath, nil); found {
		// First component must match a resource.
		if len(path) >= 1 {
			// If there are some components left, the path targets an item or an alias.

			// Shift the item id from the path components.
			var id string
			id, path = nextPathComponent(path)

			// Handle sub-resources (/resource1/id1/resource2/id2).
			if len(path) >= 1 {
				subPathComp, _ := nextPathComponent(path)
				subResourcePath := resourcePath + "." + subPathComp
				if subResource, found := index.GetResource(subResourcePath, nil); found {
					// Append the intermediate resource path.
					if err := route.ResourcePath.append(rsrc, subResource.ParentField(), id, name); err != nil {
						return err
					}
					// Recurse to match the sub-path.
					if err := findRoute(path, index, route); err != nil {
						return err
					}
				} else {
					route.ResourcePath.clear()
					return errResourceNotFound
				}
				return nil
			}

			// Handle aliases (/resource/alias or /resource1/id1/resource2/alias).
			if alias, found := rsrc.GetAlias(id); found {
				// Apply aliases query to the request.
				for key, values := range alias {
					for _, value := range values {
						route.Params.Add(key, value)
					}
				}
			} else {
				// Set the id route field.
				return route.ResourcePath.append(rsrc, "id", id, name)
			}
		}
		// Set the collection resource.
		return route.ResourcePath.append(rsrc, "", nil, name)
	}
	route.ResourcePath.clear()
	return errResourceNotFound
}

// nextPathComponent returns the next path component and the remaining path
//
// Input: /comp1/comp2/comp3
// Output: comp1, comp2/comp3
func nextPathComponent(path string) (string, string) {
	// Remove leading slash if any
	for len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	comp := path
	if i := strings.IndexByte(path, '/'); i != -1 {
		comp = path[:i]
		path = path[i+1:]
	} else {
		path = path[0:0]
	}
	return comp, path
}

// Resource returns the last resource path's resource.
func (r *RouteMatch) Resource() *resource.Resource {
	l := len(r.ResourcePath)
	if l == 0 {
		return nil
	}
	return (r.ResourcePath)[l-1].Resource
}

// ResourceID returns the last resource path's resource id value if any.
//
// If this method returns a non nil value, it means the route is an item request,
// otherwise it's a collection request.
func (r *RouteMatch) ResourceID() interface{} {
	l := len(r.ResourcePath)
	if l == 0 {
		return nil
	}
	return (r.ResourcePath)[l-1].Value
}

// Query builds a query object from the matched route
func (r *RouteMatch) Query() (*query.Query, *Error) {
	qp := queryParser{rsc: r.Resource()}
	if qp.rsc == nil {
		return nil, &Error{500, "missing resource", nil}
	}

	// Append route fields to the query
	for _, rp := range r.ResourcePath {
		if rp.Value != nil {
			qp.q.Predicate = append(qp.q.Predicate, &query.Equal{Field: rp.Field, Value: rp.Value})
		}
	}

	// Parse query string params.
	switch r.Method {
	case "DELETE":
		qp.parsePredicate(r.Params)
		qp.parseWindow(r.Params, false)
		qp.parseSort(r.Params)
	case "HEAD", "GET":
		qp.parsePredicate(r.Params)
		qp.parseWindow(r.Params, true)
		qp.parseSort(r.Params)
		qp.parseProjection(r.Params)
	case "POST", "PUT", "PATCH":
		// Allow projection to be applied on mutation responses that return
		// the mutated item.
		qp.parseProjection(r.Params)
	}

	return qp.results()
}

// Release releases the route so it can be reused.
func (r *RouteMatch) Release() {
	r.Params = nil
	r.Method = ""
	r.ResourcePath.clear()
	routePool.Put(r)
}

// queryParser is a small helper type that parses query parameters, while also
// storing any potential query issues for a combined error result.
type queryParser struct {
	q      query.Query
	issues map[string][]interface{}
	rsc    *resource.Resource
}

func (qp *queryParser) results() (*query.Query, *Error) {
	if len(qp.issues) > 0 {
		return nil, &Error{422, "URL parameters contain error(s)", qp.issues}
	}
	return &qp.q, nil
}

func (qp *queryParser) addIssue(field string, err interface{}) {
	if qp.issues == nil {
		qp.issues = map[string][]interface{}{}
	}
	qp.issues[field] = append(qp.issues[field], err)
}

func (qp *queryParser) parseProjection(params url.Values) {
	if fields := params.Get("fields"); fields != "" {
		if p, err := query.ParseProjection(fields); err != nil {
			qp.addIssue("fields", err.Error())
		} else if err := p.Validate(qp.rsc.Validator()); err != nil {
			qp.addIssue("fields", err.Error())
		} else {
			qp.q.Projection = p
		}
	}
}

func (qp *queryParser) parsePredicate(params url.Values) {
	if filters, found := params["filter"]; found {
		// If several filter parameters are present, merge them using $and
		for _, filter := range filters {
			if p, err := query.ParsePredicate(filter); err != nil {
				qp.addIssue("filter", err.Error())
			} else if err := p.Prepare(qp.rsc.Validator()); err != nil {
				qp.addIssue("filter", err.Error())
			} else {
				qp.q.Predicate = append(qp.q.Predicate, p...)
			}
		}
	}
}

func (qp *queryParser) parseSort(params url.Values) {
	if sort := params.Get("sort"); sort != "" {
		if s, err := query.ParseSort(sort); err != nil {
			qp.addIssue("sort", err.Error())
		} else if err := s.Validate(qp.rsc.Validator()); err != nil {
			qp.addIssue("sort", err.Error())
		} else {
			qp.q.Sort = s
		}
	}
}

func (qp *queryParser) parseWindow(params url.Values, allowDefaultLimit bool) {
	limit := -1
	if l, found, err := getUintParam(params, "limit"); found {
		if err != nil {
			qp.addIssue("limit", err.Error())
		} else {
			limit = l
		}
	} else if allowDefaultLimit {
		if l := qp.rsc.Conf().PaginationDefaultLimit; l > 0 {
			limit = l
		}
	}
	skip := 0
	if s, found, err := getUintParam(params, "skip"); found {
		if err != nil {
			qp.addIssue("skip", err.Error())
		} else {
			skip = s
		}
	}
	page := 1
	if p, found, err := getUintParam(params, "page"); found {
		if err != nil {
			qp.addIssue("page", err.Error())
		} else {
			page = p
		}
	}
	if page > 1 && limit <= 0 {
		qp.addIssue("limit", "required when page is set and there is no resource default")
	}

	qp.q.Window = query.Page(page, limit, skip)
}
