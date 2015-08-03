package rest

import "golang.org/x/net/context"

// itemPatch handles PATCH resquests on an item URL
//
// Reference: http://tools.ietf.org/html/rfc5789
func (r *request) itemPatch(ctx context.Context, route route) {
	var payload map[string]interface{}
	if err := r.decodePayload(&payload); err != nil {
		r.sendError(err)
		return
	}
	lookup, err := route.lookup()
	if err != nil {
		r.sendError(err)
	}
	// Get original item if any
	var original *Item
	if l, err := route.resource.handler.Find(lookup, 1, 1, ctx); err != nil {
		// If item can't be fetch, return an error
		r.sendError(err)
		return
	} else if len(l.Items) == 0 {
		r.sendError(NotFoundError)
		return
	} else {
		original = l.Items[0]
	}
	// If-Match / If-Unmodified-Since handling
	if err := r.checkIntegrityRequest(original); err != nil {
		r.sendError(err)
		return
	}
	changes, base := route.resource.schema.Prepare(payload, &original.Payload, false)
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
	item, err2 := NewItem(doc)
	if err != nil {
		r.sendError(err2)
		return
	}
	// Store the modified document by providing the orignal doc to instruct
	// handler to ensure the stored document didn't change between in the
	// interval. An PreconditionFailedError will be thrown in case of race condition
	// (i.e.: another thread modified the document between the Find() and the Store())
	if err := route.resource.handler.Update(item, original, ctx); err != nil {
		r.sendError(err)
	} else {
		r.sendItem(200, item)
	}
}
