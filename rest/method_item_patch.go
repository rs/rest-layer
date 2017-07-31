package rest

import (
	"context"
	"net/http"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema/query"
)

// itemPatch handles PATCH resquests on an item URL.
//
// Reference: http://tools.ietf.org/html/rfc5789
func itemPatch(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	var payload map[string]interface{}
	if e := decodePayload(r, &payload); e != nil {
		return e.Code, nil, e
	}
	q, e := route.Query()
	if e != nil {
		return e.Code, nil, e
	}
	// Get original item if any.
	rsrc := route.Resource()
	var original *resource.Item
	q.Window = &query.Window{Limit: 1}
	if l, err := rsrc.Find(ctx, q); err != nil {
		// If item can't be fetch, return an error.
		e = NewError(err)
		return e.Code, nil, e
	} else if len(l.Items) == 0 {
		return ErrNotFound.Code, nil, ErrNotFound
	} else {
		original = l.Items[0]
	}
	// If-Match / If-Unmodified-Since handling.
	if err := checkIntegrityRequest(r, original); err != nil {
		return err.Code, nil, err
	}
	changes, base := rsrc.Validator().Prepare(ctx, payload, &original.Payload, false)
	// Append lookup fields to base payload so it isn't caught by ReadOnly
	// (i.e.: contains id and parent resource refs if any).
	for k, v := range route.ResourcePath.Values() {
		base[k] = v
	}
	doc, errs := rsrc.Validator().Validate(changes, base)
	if len(errs) > 0 {
		return 422, nil, &Error{422, "Document contains error(s)", errs}
	}
	if id, found := doc["id"]; found && id != original.ID {
		return 422, nil, &Error{422, "Cannot change document ID", nil}
	}
	item, err := resource.NewItem(doc)
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	// Store the modified document by providing the original doc to instruct
	// handler to ensure the stored document didn't change between in the
	// interval. An ErrPreconditionFailed will be thrown in case of race
	// condition (i.e.: another thread modified the document between the Find()
	// and the Store()).
	if err = rsrc.Update(ctx, item, original); err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	// Evaluate projection so response gets the same format as read requests.
	item.Payload, err = q.Projection.Eval(ctx, item.Payload, restResource{rsrc})
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	return 200, nil, item
}
