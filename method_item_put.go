package rest

import "golang.org/x/net/context"

// itemPut handles PUT resquests on an item URL
//
// Reference: http://tools.ietf.org/html/rfc2616#section-9.6
func (r *request) itemPut(ctx context.Context, route route) {
	var payload map[string]interface{}
	if err := r.decodePayload(&payload); err != nil {
		r.sendError(err)
		return
	}
	lookup, err := route.lookup()
	if err != nil {
		r.sendError(err)
	}
	// Fetch original item if exist (PUT can be used to create a document with a manual id)
	var original *Item
	if l, err := route.resource.handler.Find(lookup, 1, 1, ctx); err != nil && err != NotFoundError {
		r.sendError(err)
		return
	} else if len(l.Items) == 1 {
		original = l.Items[0]
	}
	// Check if method is allowed based
	mode := Create
	if original != nil {
		// If original is found, the mode is replace rather than create
		mode = Replace
	}
	if !route.resource.conf.isModeAllowed(mode) {
		r.sendError(&Error{405, "Invalid method", nil})
		return
	}
	// If-Match / If-Unmodified-Since handling
	if err := r.checkIntegrityRequest(original); err != nil {
		r.sendError(err)
		return
	}
	status := 200
	var changes map[string]interface{}
	var base map[string]interface{}
	if original == nil {
		// PUT used to create a new document
		changes, base = route.resource.schema.Prepare(payload, nil, false)
		status = 201
	} else {
		// PUT used to replace an existing document
		changes, base = route.resource.schema.Prepare(payload, &original.Payload, true)
	}
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
	if original != nil {
		if id, found := doc["id"]; found && id != original.ID {
			r.sendError(&Error{422, "Cannot change document ID", nil})
			return
		}
	}
	item, err2 := NewItem(doc)
	if err != nil {
		r.sendError(err2)
		return
	}
	// If we have an original item, pass it to the handler so we make sure
	// we are still replacing the same version of the object as handler is
	// supposed check the original etag before storing when an original object
	// is provided.
	if original != nil {
		if err := route.resource.handler.Update(item, original, ctx); err != nil {
			r.sendError(err)
			return
		}
	} else {
		if err := route.resource.handler.Insert([]*Item{item}, ctx); err != nil {
			r.sendError(err)
			return
		}
	}
	r.sendItem(status, item)
}
