package rest

import (
	"net/http"

	"github.com/rs/rest-layer/resource"
	"golang.org/x/net/context"
)

// itemPut handles PUT resquests on an item URL
//
// Reference: http://tools.ietf.org/html/rfc2616#section-9.6
func (r *request) itemPut(ctx context.Context, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	var payload map[string]interface{}
	if e := r.decodePayload(&payload); e != nil {
		return e.Code, nil, e
	}
	lookup, e := route.Lookup()
	if e != nil {
		return e.Code, nil, e
	}
	rsrc := route.Resource()
	// Fetch original item if exist (PUT can be used to create a document with a manual id)
	var original *resource.Item
	if l, err := rsrc.Find(ctx, lookup, 1, 1); err != nil && err != ErrNotFound {
		e = NewError(err)
		return e.Code, nil, e
	} else if len(l.Items) == 1 {
		original = l.Items[0]
	}
	// Check if method is allowed based
	mode := resource.Create
	if original != nil {
		// If original is found, the mode is replace rather than create
		mode = resource.Replace
	}
	if !rsrc.Conf().IsModeAllowed(mode) {
		return 405, nil, &Error{405, "Invalid method", nil}
	}
	// If-Match / If-Unmodified-Since handling
	if err := r.checkIntegrityRequest(original); err != nil {
		return err.Code, nil, err
	}
	status = 200
	var changes map[string]interface{}
	var base map[string]interface{}
	if original == nil {
		// PUT used to create a new document
		changes, base = rsrc.Validator().Prepare(payload, nil, false)
		status = 201
	} else {
		// PUT used to replace an existing document
		changes, base = rsrc.Validator().Prepare(payload, &original.Payload, true)
	}
	// Append lookup fields to base payload so it isn't caught by ReadOnly
	// (i.e.: contains id and parent resource refs if any)
	route.applyFields(base)
	doc, errs := rsrc.Validator().Validate(changes, base)
	if len(errs) > 0 {
		return 422, nil, &Error{422, "Document contains error(s)", errs}
	}
	// Check that fields with the Reference validator reference an existing object
	if err := r.checkReferences(ctx, doc, rsrc.Validator()); err != nil {
		return err.Code, nil, err
	}
	if original != nil {
		if id, found := doc["id"]; found && id != original.ID {
			return 422, nil, &Error{422, "Cannot change document ID", nil}
		}
	}
	item, err := resource.NewItem(doc)
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	// If we have an original item, pass it to the handler so we make sure
	// we are still replacing the same version of the object as handler is
	// supposed check the original etag before storing when an original object
	// is provided.
	if original != nil {
		if err := rsrc.Update(ctx, item, original); err != nil {
			e = NewError(err)
			return e.Code, nil, e
		}
	} else {
		if err := rsrc.Insert(ctx, []*resource.Item{item}); err != nil {
			e = NewError(err)
			return e.Code, nil, e
		}
	}
	return status, nil, item
}
