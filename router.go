package rest

import "strings"

// route recursively route a (sub)resource request
func (r *requestHandler) route(path string, lookup *Lookup, resources map[string]*subResource) {
	// Split the path into path components
	c := strings.Split(path, "/")

	if len(c) == 0 {
		return
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
					lookup.Fields["id"] = id
					l, err := resource.handler.Find(lookup, 1, 1, r.ctx)
					if err != nil {
						r.sendError(err)
						return
					} else if len(l.Items) == 0 {
						r.sendError(NotFoundError)
						return
					}
					// Move item's current id to the sub resource's filter
					delete(lookup.Fields, "id")
					lookup.Fields[sub.field] = id
					path = strings.Join(c, "/")
					r.route(path, lookup, resource.resources)
				} else {
					r.sendError(&Error{404, "Resource not found", nil})
				}
				return
			}

			// Handle aliases (/resource/alias or /resource1/id1/resource2/alias)
			if alias, found := resource.aliases[id]; found {
				// Apply aliases query to the request
				q := r.req.URL.Query()
				for key, values := range alias {
					for _, value := range values {
						q.Add(key, value)
					}
				}
				r.handleResourceRequest(resource, lookup)
				return
			}

			// Set the id filter
			lookup.Fields["id"] = id

			r.handleItemRequest(resource, lookup)
		} else {
			r.handleResourceRequest(resource, lookup)
		}
	} else {
		r.sendError(&Error{404, "Resource not found", nil})
	}
}

func (r *requestHandler) handleResourceRequest(resource *Resource, lookup *Lookup) {
	switch r.req.Method {
	case "OPTIONS":
		methods := []string{}
		if resource.conf.isModeAllowed(List) {
			methods = append(methods, "HEAD", "GET")
		}
		if resource.conf.isModeAllowed(Create) {
			methods = append(methods, "POST")
		}
		if resource.conf.isModeAllowed(Clear) {
			methods = append(methods, "DELETE")
		}
		r.res.Header().Set("Allow", strings.Join(methods, ", "))
	case "HEAD", "GET":
		if !resource.conf.isModeAllowed(List) {
			r.sendError(InvalidMethodError)
			return
		}
		r.handleListRequestGET(lookup, resource)
	case "POST":
		if !resource.conf.isModeAllowed(Create) {
			r.sendError(InvalidMethodError)
			return
		}
		r.handleListRequestPOST(lookup, resource)
	case "DELETE":
		if !resource.conf.isModeAllowed(Clear) {
			r.sendError(InvalidMethodError)
			return
		}
		r.handleListRequestDELETE(lookup, resource)
	default:
		r.sendError(InvalidMethodError)
	}
}

func (r *requestHandler) handleItemRequest(resource *Resource, lookup *Lookup) {
	switch r.req.Method {
	case "OPTIONS":
		methods := []string{}
		if resource.conf.isModeAllowed(Read) {
			methods = append(methods, "HEAD", "GET")
		}
		if resource.conf.isModeAllowed(Create) || resource.conf.isModeAllowed(Replace) {
			methods = append(methods, "PUT")
		}
		if resource.conf.isModeAllowed(Update) {
			methods = append(methods, "PATCH")
			// See http://tools.ietf.org/html/rfc5789#section-3
			r.res.Header().Set("Allow-Patch", "application/json")
		}
		if resource.conf.isModeAllowed(Update) {
			methods = append(methods, "DELETE")
		}
		r.res.Header().Set("Allow", strings.Join(methods, ", "))
	case "HEAD", "GET":
		if !resource.conf.isModeAllowed(Read) {
			r.sendError(InvalidMethodError)
			return
		}
		r.handleItemRequestGET(lookup, resource)
	case "PUT":
		if !resource.conf.isModeAllowed(Create) && !resource.conf.isModeAllowed(Replace) {
			r.sendError(InvalidMethodError)
			return
		}
		r.handleItemRequestPUT(lookup, resource)
	case "PATCH":
		if !resource.conf.isModeAllowed(Update) {
			r.sendError(InvalidMethodError)
			return
		}
		r.handleItemRequestPATCH(lookup, resource)
	case "DELETE":
		if !resource.conf.isModeAllowed(Delete) {
			r.sendError(InvalidMethodError)
			return
		}
		r.handleItemRequestDELETE(lookup, resource)
	default:
		r.sendError(InvalidMethodError)
	}
}
