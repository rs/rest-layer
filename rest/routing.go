package rest

import (
	"context"
	"fmt"
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
	if req.URL.RawQuery != "" {
		route.Params = req.URL.Query()
	}
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
	q := &query.Query{}
	// Append route fields to the query
	for _, rp := range r.ResourcePath {
		if rp.Value != nil {
			q.Predicate = append(q.Predicate, query.Equal{Field: rp.Field, Value: rp.Value})
		}
	}
	rsc := r.Resource()
	if rsc == nil {
		return nil, &Error{500, "missing resource", nil}
	}
	// Parse query string params.
	switch r.Method {
	case "HEAD", "GET":
		limit := -1
		if l, found, err := getUintParam(r.Params, "limit"); found {
			if err != nil {
				return nil, &Error{422, err.Error(), nil}
			}
			limit = l
		} else if l := rsc.Conf().PaginationDefaultLimit; l > 0 {
			limit = l
		}
		skip := 0
		if s, found, err := getUintParam(r.Params, "skip"); found {
			if err != nil {
				return nil, &Error{422, err.Error(), nil}
			}
			skip = s
		}
		page := 1
		if p, found, err := getUintParam(r.Params, "page"); found {
			if err != nil {
				return nil, &Error{422, err.Error(), nil}
			}
			page = p
		}
		if page > 1 && limit <= 0 {
			return nil, &Error{422, "Cannot use `page' parameter with no `limit' parameter on a resource with no default pagination size", nil}
		}
		q.Window = query.Page(page, limit, skip)
	}
	if sort := r.Params.Get("sort"); sort != "" {
		s, err := query.ParseSort(sort)
		if err == nil {
			err = s.Validate(rsc.Validator())
		}
		if err != nil {
			return nil, &Error{422, fmt.Sprintf("Invalid `sort` parameter: %s", err), nil}
		}
		q.Sort = s
	}
	if filters, found := r.Params["filter"]; found {
		// If several filter parameters are present, merge them using $and
		for _, filter := range filters {
			p, err := query.ParsePredicate(filter)
			if err == nil {
				err = p.Validate(rsc.Validator())
			}
			if err != nil {
				return nil, &Error{422, fmt.Sprintf("Invalid `filter` parameter: %s", err), nil}
			}
			q.Predicate = append(q.Predicate, p...)
		}
	}
	if fields := r.Params.Get("fields"); fields != "" {
		p, err := query.ParseProjection(fields)
		if err == nil {
			err = p.Validate(rsc.Validator())
		}
		if err != nil {
			return nil, &Error{422, fmt.Sprintf("Invalid `fields` parameter: %s", err), nil}
		}
		q.Projection = p
	}
	return q, nil
}

// Release releases the route so it can be reused.
func (r *RouteMatch) Release() {
	r.Params = nil
	r.Method = ""
	r.ResourcePath.clear()
	routePool.Put(r)
}
