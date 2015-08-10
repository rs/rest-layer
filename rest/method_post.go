package rest

import (
	"fmt"
	"net/http"

	"github.com/rs/rest-layer/resource"
	"golang.org/x/net/context"
)

// listPost handles POST resquests on a resource URL
func (r *request) listPost(ctx context.Context, route RouteMatch) (status int, headers http.Header, body interface{}) {
	var payload map[string]interface{}
	if e := r.decodePayload(&payload); e != nil {
		return e.Code, nil, e
	}
	rsrc := route.Resource()
	changes, base := rsrc.Validator().Prepare(payload, nil, false)
	// Append lookup fields to base payload so it isn't caught by ReadOnly
	// (i.e.: contains id and parent resource refs if any)
	route.applyFields(base)
	doc, errs := rsrc.Validator().Validate(changes, base)
	if len(errs) > 0 {
		return 422, nil, &Error{422, "Document contains error(s)", errs}
	}
	// Check that fields with the Reference validator reference an existing object
	if err := r.checkReferences(ctx, doc, rsrc.Validator()); err != nil {
		e := NewError(err)
		return e.Code, nil, e
	}
	item, err := resource.NewItem(doc)
	if err != nil {
		e := NewError(err)
		return e.Code, nil, e
	}
	// TODO: add support for batch insert
	if err := rsrc.Insert(ctx, []*resource.Item{item}); err != nil {
		e := NewError(err)
		return e.Code, nil, e
	}
	// See https://www.subbu.org/blog/2008/10/location-vs-content-location
	headers = http.Header{}
	headers.Set("Content-Location", fmt.Sprintf("/%s/%s", r.req.URL.Path, item.ID))
	return 201, headers, item
}
