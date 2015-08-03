package rest

import (
	"fmt"

	"golang.org/x/net/context"
)

// listPost handles POST resquests on a resource URL
func (r *request) listPost(ctx context.Context, route route) {
	var payload map[string]interface{}
	if err := r.decodePayload(&payload); err != nil {
		r.sendError(err)
		return
	}
	changes, base := route.resource.schema.Prepare(payload, nil, false)
	// Append lookup fields to base payload so it isn't caught by ReadOnly
	// (i.e.: contains id and parent resource refs if any)
	route.applyFields(base)
	doc, errs := route.resource.schema.Validate(changes, base)
	if len(errs) > 0 {
		r.sendError(&Error{422, "Document contains error(s)", errs})
		return
	}
	// Check that fields with the Reference validator reference an existing object
	if err := r.checkReferences(ctx, doc, route.resource.schema); err != nil {
		r.sendError(err)
		return
	}
	item, err := NewItem(doc)
	if err != nil {
		r.sendError(err)
		return
	}
	// TODO: add support for batch insert
	if err := route.resource.handler.Insert([]*Item{item}, ctx); err != nil {
		r.sendError(err)
		return
	}
	// See https://www.subbu.org/blog/2008/10/location-vs-content-location
	r.res.Header().Set("Content-Location", fmt.Sprintf("/%s/%s", r.req.URL.Path, item.ID))
	r.sendItem(201, item)
}
