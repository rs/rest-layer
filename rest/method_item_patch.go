package rest

import (
	"net/http"

	"github.com/rs/rest-layer/resource"
	"golang.org/x/net/context"
)

// itemPatch handles PATCH resquests on an item URL
//
// Reference: http://tools.ietf.org/html/rfc5789
func (r *request) itemPatch(ctx context.Context, route RouteMatch) (status int, headers http.Header, body interface{}) {
	var payload map[string]interface{}
	if e := r.decodePayload(&payload); e != nil {
		return e.Code, nil, e
	}
	lookup, e := route.Lookup()
	if e != nil {
		return e.Code, nil, e
	}
	// Get original item if any
	rsrc := route.Resource()
	var original *resource.Item
	if l, err := rsrc.Find(ctx, lookup, 1, 1); err != nil {
		// If item can't be fetch, return an error
		e = NewError(err)
		return e.Code, nil, e
	} else if len(l.Items) == 0 {
		return ErrNotFound.Code, nil, ErrNotFound
	} else {
		original = l.Items[0]
	}
	// If-Match / If-Unmodified-Since handling
	if err := r.checkIntegrityRequest(original); err != nil {
		return e.Code, nil, e
	}
	changes, base := rsrc.Validator().Prepare(payload, &original.Payload, false)
	// Append lookup fields to base payload so it isn't caught by ReadOnly
	// (i.e.: contains id and parent resource refs if any)
	route.applyFields(base)
	doc, errs := rsrc.Validator().Validate(changes, base)
	if len(errs) > 0 {
		return 422, nil, &Error{422, "Document contains error(s)", errs}
	}
	// Check that fields with the Reference validator reference an existing object
	if e := r.checkReferences(ctx, doc, rsrc.Validator()); e != nil {
		return e.Code, nil, e
	}
	item, err := resource.NewItem(doc)
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	// Store the modified document by providing the orignal doc to instruct
	// handler to ensure the stored document didn't change between in the
	// interval. An ErrPreconditionFailed will be thrown in case of race condition
	// (i.e.: another thread modified the document between the Find() and the Store())
	if err := rsrc.Update(ctx, item, original); err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	return 200, nil, item
}
