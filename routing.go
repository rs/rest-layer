package rest

import (
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"
)

// ResourceRouter is an interface to a type able to find a Resource from a resource path
type ResourceRouter interface {
	// Bind a new resource at the "name" endpoint
	Bind(name string, s *Resource) *Resource
	// GetResource retrives a given resource and its parent field identifier by it's path.
	// For instance if a resource user has a sub-resource posts,
	// a users.posts path can be use to retrieve the posts resource.
	GetResource(path string) (*Resource, string, bool)
	// FindRoute returns the REST route for the given request
	FindRoute(ctx context.Context, req *http.Request) (RouteMatch, *Error)
}

// RouteMatch represent a REST request's matched resource with the method to apply and its parameters
type RouteMatch struct {
	// Method is the HTTP method used on the resource.
	Method string
	// ResourcePath is the list of intermediate resources followed by the targetted resource.
	ResourcePath ResourcePath
	// Params is the list of client provided parameters (thru query-string or alias).
	Params url.Values
}

// ResourcePath is the list of ResourcePathComponent leading to the requested resource
type ResourcePath []ResourcePathComponent

// ResourcePathComponent represents the path of resource and sub-resources of a given request's resource
type ResourcePathComponent struct {
	// Name is the endpoint name used to bind the resource
	Name string
	// Field is the resource's field used to filter targetted resource
	Field string
	// Value holds the resource's id value
	Value interface{}
	// Resource references the resource
	Resource *Resource
}

type key int

const (
	routeKey key = iota
	routerKey
)

func contextWithRoute(ctx context.Context, route *RouteMatch) context.Context {
	return context.WithValue(ctx, routeKey, route)
}

func contextWithRouter(ctx context.Context, router ResourceRouter) context.Context {
	return context.WithValue(ctx, routerKey, router)
}

// RouteFromContext extracts the matched route from the given net/context
func RouteFromContext(ctx context.Context) (*RouteMatch, bool) {
	route, ok := ctx.Value(routeKey).(*RouteMatch)
	return route, ok
}

// RouterFromContext extracts the router from the given net/context
func RouterFromContext(ctx context.Context) (ResourceRouter, bool) {
	router, ok := ctx.Value(routerKey).(ResourceRouter)
	return router, ok
}

// FindRoute returns the REST route for the given request
func (r *rootResource) FindRoute(ctx context.Context, req *http.Request) (RouteMatch, *Error) {
	route := RouteMatch{
		Method:       req.Method,
		ResourcePath: ResourcePath{},
		Params:       req.URL.Query(),
	}
	err := findRoute(ctx, req.URL.Path, r, &route)
	return route, err
}

// findRoute recursively route a (sub)resource request
func findRoute(ctx context.Context, path string, router ResourceRouter, route *RouteMatch) *Error {
	// Split the path into path components
	c := strings.Split(strings.Trim(path, "/"), "/")

	// Shift the resource name from the path components
	name, c := c[0], c[1:]

	resourcePath := name
	if prefix := route.ResourcePath.Path(); prefix != "" {
		resourcePath = strings.Join([]string{prefix, name}, ".")
	}

	// First component must match a resource
	if resource, _, found := router.GetResource(resourcePath); found {
		rp := ResourcePathComponent{
			Name:     name,
			Resource: resource,
		}
		if len(c) >= 1 {
			// If there are some components left, the path targets an item or an alias

			// Shift the item id from the path components
			var id string
			id, c = c[0], c[1:]

			// Handle sub-resources (/resource1/id1/resource2/id2)
			if len(c) >= 1 {
				subResourcePath := strings.Join([]string{resourcePath, c[0]}, ".")
				if _, field, found := router.GetResource(subResourcePath); found {
					// Check if the current (intermediate) item exists before going farther
					l := newLookup()
					for _, rp := range route.ResourcePath {
						if rp.Value != nil {
							l.filter[rp.Field] = rp.Value
						}
					}
					l.filter["id"] = id
					list, err := resource.handler.Find(ctx, l, 1, 1)
					if err != nil {
						return err
					} else if len(list.Items) == 0 {
						return NotFoundError
					}
					rp.Field = field
					rp.Value = id
					route.ResourcePath = append(route.ResourcePath, rp)
					// Recurse to match the sub-path
					path = strings.Join(c, "/")
					if err := findRoute(ctx, path, router, route); err != nil {
						return err
					}
				} else {
					route.ResourcePath = ResourcePath{}
					return &Error{404, "Resource Not Found", nil}
				}
				return nil
			}

			// Handle aliases (/resource/alias or /resource1/id1/resource2/alias)
			if alias, found := resource.aliases[id]; found {
				// Apply aliases query to the request
				for key, values := range alias {
					for _, value := range values {
						route.Params.Add(key, value)
					}
				}
			} else {
				// Set the id route field
				rp.Field = "id"
				rp.Value = id
			}
		}
		route.ResourcePath = append(route.ResourcePath, rp)
		return nil
	}
	route.ResourcePath = ResourcePath{}
	return &Error{404, "Resource Not Found", nil}
}

// Path returns the path to the resource to be used with ResourceRouter.GetResource
func (p ResourcePath) Path() string {
	path := []string{}
	for _, c := range p {
		path = append(path, c.Name)
	}
	return strings.Join(path, ".")
}

// Resource returns the last resource path's resource
func (r RouteMatch) Resource() *Resource {
	l := len(r.ResourcePath)
	if l == 0 {
		return nil
	}
	return r.ResourcePath[l-1].Resource
}

// ResourceID returns the last resource path's resource id value if any.
//
// If this method returns a non nil value, it means the route is an item request,
// otherwise it's a collection request.
func (r RouteMatch) ResourceID() interface{} {
	l := len(r.ResourcePath)
	if l == 0 {
		return nil
	}
	return r.ResourcePath[l-1].Value
}

// Lookup builds a Lookup object from the matched route
func (r RouteMatch) Lookup() (Lookup, *Error) {
	l := newLookup()
	if sort := r.Params.Get("sort"); sort != "" {
		if err := l.setSort(sort, r.Resource().schema); err != nil {
			return nil, &Error{422, "Invalid `sort` paramter", nil}
		}
	}
	// TODO: Handle multiple filter param
	if filter := r.Params.Get("filter"); filter != "" {
		if err := l.setFilter(filter, r.Resource().schema); err != nil {
			return nil, &Error{422, "Invalid `filter` parameter", nil}
		}
	}
	// Append route fields to the query
	for _, rp := range r.ResourcePath {
		// TODO: handle collisions
		if rp.Value != nil {
			l.filter[rp.Field] = rp.Value
		}
	}
	return l, nil
}

// applyFields appends lookup fields to a payload
func (r RouteMatch) applyFields(payload map[string]interface{}) {
	for _, rp := range r.ResourcePath {
		if rp.Value != nil {
			payload[rp.Field] = rp.Value
		}
	}
}
