package rest

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/rest-layer/schema"
	"golang.org/x/net/context"
)

// route represent a REST request's matched resource with the method to apply and its parameters
type route struct {
	method   string
	resource *Resource
	fields   map[string]interface{}
	params   url.Values
}

// findRoute returns the REST route for the given request
func (r *RootResource) findRoute(ctx context.Context, req *http.Request) (route, *Error) {
	route := route{
		method: req.Method,
		fields: map[string]interface{}{},
		params: req.URL.Query(),
	}
	err := findRoute(ctx, req.URL.Path, r.resources, &route)
	return route, err
}

// findRoute recursively route a (sub)resource request
func findRoute(ctx context.Context, path string, resources map[string]*subResource, route *route) *Error {
	// Split the path into path components
	c := strings.Split(strings.Trim(path, "/"), "/")

	if len(c) == 0 {
		return nil
	}

	// Shift the resource name from the path components
	name, c := c[0], c[1:]

	// First component must match a resource
	if sr, found := resources[name]; found {
		resource := sr.resource
		if len(c) >= 1 {
			// If there are some components left, the path targets an item or an alias

			// Shift the item id from the path components
			var id string
			id, c = c[0], c[1:]

			// Handle sub-resources (/resource1/id1/resource2/id2)
			if len(c) >= 1 {
				subName := c[0]
				if sub, found := resource.resources[subName]; found {
					// Check if the item exists before going farther
					q := schema.NewQuery(route.fields)
					q["id"] = id
					l, err := resource.handler.Find(&Lookup{Filter: q}, 1, 1, ctx)
					if err != nil {
						return err
					} else if len(l.Items) == 0 {
						return NotFoundError
					}
					// Move item's current id to the sub resource's filter
					route.fields[sub.field] = id
					route.resource = resource
					path = strings.Join(c, "/")
					if err := findRoute(ctx, path, resource.resources, route); err != nil {
						return err
					}
				} else {
					return &Error{404, "Resource Not Found", nil}
				}
				return nil
			}

			// Handle aliases (/resource/alias or /resource1/id1/resource2/alias)
			if alias, found := resource.aliases[id]; found {
				// Apply aliases query to the request
				for key, values := range alias {
					for _, value := range values {
						route.params.Add(key, value)
					}
				}
			} else {
				// Set the id route field
				route.fields["id"] = id
			}
		}
		route.resource = resource
		return nil
	}
	return &Error{404, "Resource Not Found", nil}
}

// lookup builds a Lookup object from the current route
func (r route) lookup() (*Lookup, *Error) {
	lookup := NewLookup()
	if sort := r.params.Get("sort"); sort != "" {
		if err := lookup.SetSort(sort, r.resource.schema); err != nil {
			return nil, &Error{422, "Invalid `sort` paramter", nil}
		}
	}
	// TODO: Handle multiple filter param
	if filter := r.params.Get("filter"); filter != "" {
		if err := lookup.SetFilter(filter, r.resource.schema); err != nil {
			return nil, &Error{422, "Invalid `filter` parameter", nil}
		}
	}
	// Append route fields to the query
	for field, value := range r.fields {
		// TODO: handle collisions
		lookup.Filter[field] = value
	}
	return lookup, nil
}

// applyFields appends lookup fields to a payload
func (r route) applyFields(payload map[string]interface{}) {
	for field, value := range r.fields {
		payload[field] = value
	}
}
