package rest

import (
	"net/http"

	"golang.org/x/net/context"
)

// itemPatch handles PATCH resquests on an item URL
//
// Reference: http://tools.ietf.org/html/rfc5789
func (r *request) itemPatch(ctx context.Context, route RouteMatch) (status int, headers http.Header, body interface{}) {
	var payload map[string]interface{}
	if err := r.decodePayload(&payload); err != nil {
		return err.Code, nil, err
	}
	lookup, err := route.Lookup()
	if err != nil {
		return err.Code, nil, err
	}
	// Get original item if any
	resource := route.Resource()
	var original *Item
	if l, err := resource.handler.Find(ctx, lookup, 1, 1); err != nil {
		// If item can't be fetch, return an error
		return err.Code, nil, err
	} else if len(l.Items) == 0 {
		return NotFoundError.Code, nil, NotFoundError
	} else {
		original = l.Items[0]
	}
	// If-Match / If-Unmodified-Since handling
	if err := r.checkIntegrityRequest(original); err != nil {
		return err.Code, nil, err
	}
	changes, base := resource.schema.Prepare(payload, &original.Payload, false)
	// Append lookup fields to base payload so it isn't caught by ReadOnly
	// (i.e.: contains id and parent resource refs if any)
	route.applyFields(base)
	doc, errs := resource.schema.Validate(changes, base)
	if len(errs) > 0 {
		return 422, nil, &Error{422, "Document contains error(s)", errs}
	}
	// Check that fields with the Reference validator reference an existing object
	if err := r.checkReferences(ctx, doc, resource.schema); err != nil {
		return err.Code, nil, err
	}
	item, e := NewItem(doc)
	if e != nil {
		return 500, nil, e
	}
	// Store the modified document by providing the orignal doc to instruct
	// handler to ensure the stored document didn't change between in the
	// interval. An PreconditionFailedError will be thrown in case of race condition
	// (i.e.: another thread modified the document between the Find() and the Store())
	if err := resource.handler.Update(ctx, item, original); err != nil {
		return err.Code, nil, err
	}
	return 200, nil, item
}
