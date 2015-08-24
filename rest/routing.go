package rest

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"golang.org/x/net/context"
)

// RouteMatch represent a REST request's matched resource with the method to apply and its parameters
type RouteMatch struct {
	// Method is the HTTP method used on the resource.
	Method string
	// ResourcePath is the list of intermediate resources followed by the targetted resource.
	// Each intermediate resource mutch match all the previous resource components of this path
	// and newly created resources will have their corresponding fields filled with resource
	// path information (resource.field => resource.value).
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
	Resource *resource.Resource
}

type key int

const (
	routeKey key = iota
	indexKey
)

func contextWithRoute(ctx context.Context, route *RouteMatch) context.Context {
	return context.WithValue(ctx, routeKey, route)
}

func contextWithIndex(ctx context.Context, index resource.Index) context.Context {
	return context.WithValue(ctx, indexKey, index)
}

// RouteFromContext extracts the matched route from the given net/context
func RouteFromContext(ctx context.Context) (*RouteMatch, bool) {
	route, ok := ctx.Value(routeKey).(*RouteMatch)
	return route, ok
}

// IndexFromContext extracts the router from the given net/context
func IndexFromContext(ctx context.Context) (resource.Index, bool) {
	index, ok := ctx.Value(indexKey).(resource.Index)
	return index, ok
}

// FindRoute returns the REST route for the given request
func FindRoute(ctx context.Context, index resource.Index, req *http.Request) (RouteMatch, *Error) {
	route := RouteMatch{
		Method:       req.Method,
		ResourcePath: ResourcePath{},
		Params:       req.URL.Query(),
	}
	err := findRoute(ctx, req.URL.Path, index, &route)
	return route, err
}

// findRoute recursively route a (sub)resource request
func findRoute(ctx context.Context, path string, index resource.Index, route *RouteMatch) *Error {
	// Split the path into path components
	c := strings.Split(strings.Trim(path, "/"), "/")

	// Shift the resource name from the path components
	name, c := c[0], c[1:]

	resourcePath := name
	if prefix := route.ResourcePath.Path(); prefix != "" {
		resourcePath = strings.Join([]string{prefix, name}, ".")
	}

	// First component must match a resource
	if rsrc, _, found := index.GetResource(resourcePath); found {
		rp := ResourcePathComponent{
			Name:     name,
			Resource: rsrc,
		}
		if len(c) >= 1 {
			// If there are some components left, the path targets an item or an alias

			// Shift the item id from the path components
			var id string
			id, c = c[0], c[1:]

			// Handle sub-resources (/resource1/id1/resource2/id2)
			if len(c) >= 1 {
				subResourcePath := strings.Join([]string{resourcePath, c[0]}, ".")
				if _, field, found := index.GetResource(subResourcePath); found {
					rp.Field = field
					rp.Value = id
					route.ResourcePath = append(route.ResourcePath, rp)
					// Recurse to match the sub-path
					path = strings.Join(c, "/")
					if err := findRoute(ctx, path, index, route); err != nil {
						return err
					}
				} else {
					route.ResourcePath = ResourcePath{}
					return &Error{404, "Resource Not Found", nil}
				}
				return nil
			}

			// Handle aliases (/resource/alias or /resource1/id1/resource2/alias)
			if alias, found := rsrc.GetAlias(id); found {
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

// PrependResourcePath add the given resource using the provided field and value as a
// "ghost" resource prefix to the resource path.
//
// The effect will be a 404 error if the doesn't have an item with the id matching to the
// provided value.
//
// This will also require that all subsequent resources in the path have this resource's
// "value" set on their "field" field.
//
// Finaly, all created resources at this path will also have this field and value set by default.
func (r *RouteMatch) PrependResourcePath(rsrc *resource.Resource, field string, value interface{}) {
	rp := ResourcePathComponent{
		Field:    field,
		Value:    value,
		Resource: rsrc,
	}
	// Prepent the resource path with the user resource
	r.ResourcePath = append(ResourcePath{rp}, r.ResourcePath...)
}

// ParentsExist checks if the each intermediate parents in the path exist and
// return either a ErrNotFound or an error returned by on of the intermediate
// resource.
func (p ResourcePath) ParentsExist(ctx context.Context) error {
	// First we check that we have no field conflict on the path (i.e.: two path
	// components defining the same field with a different value)
	fields := map[string]interface{}{}
	for _, rp := range p {
		if val, found := fields[rp.Field]; found && val != rp.Value {
			return &Error{404, "Resource Path Conflict", nil}
		}
		fields[rp.Field] = rp.Value
	}

	// Check parents existence
	parents := len(p) - 1
	q := schema.Query{}
	c := make(chan error, parents)
	for _, rp := range p[:parents] {
		if rp.Value == nil {
			continue
		}
		// Create a lookup with the parent path fields + the current path id
		l := resource.NewLookup()
		lq := append(q[:], schema.Equal{Field: "id", Value: rp.Value})
		l.AddQuery(lq)
		// Execute all intermediate checkes in concurence
		go func() {
			// Check if the resource exists
			list, err := rp.Resource.Find(ctx, l, 1, 1)
			if err != nil {
				c <- err
			} else if len(list.Items) == 0 {
				c <- &Error{404, "Parent Resource Not Found", nil}
			} else {
				c <- nil
			}
		}()
		// Push the resource field=value for the next hops
		q = append(q, schema.Equal{Field: rp.Field, Value: rp.Value})
	}
	// Fail on first error
	for i := 0; i < parents; i++ {
		if err := <-c; err != nil {
			return err
		}
	}
	return nil
}

// Path returns the path to the resource to be used with resource.Root.GetResource
func (p ResourcePath) Path() string {
	path := []string{}
	for _, c := range p {
		if c.Name != "" {
			path = append(path, c.Name)
		}
	}
	return strings.Join(path, ".")
}

// Resource returns the last resource path's resource
func (r RouteMatch) Resource() *resource.Resource {
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
func (r RouteMatch) Lookup() (*resource.Lookup, *Error) {
	l := resource.NewLookup()
	// Append route fields to the query
	for _, rp := range r.ResourcePath {
		if rp.Value != nil {
			l.AddQuery(schema.Query{schema.Equal{Field: rp.Field, Value: rp.Value}})
		}
	}
	// Parse query string params
	if sort := r.Params.Get("sort"); sort != "" {
		if err := l.SetSort(sort, r.Resource().Validator()); err != nil {
			return nil, &Error{422, fmt.Sprintf("Invalid `sort` paramter: %s", err), nil}
		}
	}
	if filters, found := r.Params["filter"]; found {
		// If several filter parameters are present, merge them using $and (see lookup.addFilter)
		for _, filter := range filters {
			if err := l.AddFilter(filter, r.Resource().Validator()); err != nil {
				return nil, &Error{422, fmt.Sprintf("Invalid `filter` parameter: %s", err), nil}
			}
		}
	}
	if fields := r.Params.Get("fields"); fields != "" {
		if err := l.SetSelector(fields, r.Resource()); err != nil {
			return nil, &Error{422, fmt.Sprintf("Invalid `fields` paramter: %s", err), nil}
		}
	}
	return l, nil
}

// applyFields appends every element of the resource path to a payload
func (r RouteMatch) applyFields(payload map[string]interface{}) {
	for _, rp := range r.ResourcePath {
		if _, found := payload[rp.Field]; !found && rp.Value != nil {
			payload[rp.Field] = rp.Value
		}
	}
}
