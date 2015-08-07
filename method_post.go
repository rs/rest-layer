package rest

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"
)

// listPost handles POST resquests on a resource URL
func (r *request) listPost(ctx context.Context, route RouteMatch) (status int, headers http.Header, body interface{}) {
	var payload map[string]interface{}
	if err := r.decodePayload(&payload); err != nil {
		return err.Code, nil, err
	}
	changes, base := route.Resource().schema.Prepare(payload, nil, false)
	// Append lookup fields to base payload so it isn't caught by ReadOnly
	// (i.e.: contains id and parent resource refs if any)
	route.applyFields(base)
	doc, errs := route.Resource().schema.Validate(changes, base)
	if len(errs) > 0 {
		return 422, nil, &Error{422, "Document contains error(s)", errs}
	}
	// Check that fields with the Reference validator reference an existing object
	if err := r.checkReferences(ctx, doc, route.Resource().schema); err != nil {
		return err.Code, nil, err
	}
	item, err := NewItem(doc)
	if err != nil {
		return 500, nil, err
	}
	// TODO: add support for batch insert
	if err := route.Resource().handler.Insert(ctx, []*Item{item}); err != nil {
		return err.Code, nil, err
	}
	// See https://www.subbu.org/blog/2008/10/location-vs-content-location
	headers = http.Header{}
	headers.Set("Content-Location", fmt.Sprintf("/%s/%s", r.req.URL.Path, item.ID))
	return 201, headers, item
}
