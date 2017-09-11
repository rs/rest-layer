package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
)

// listPost handles POST resquests on a resource URL.
func listPost(ctx context.Context, r *http.Request, route *RouteMatch) (status int, headers http.Header, body interface{}) {
	q, e := route.Query()
	if e != nil {
		return e.Code, nil, e
	}
	var payload map[string]interface{}
	if e = decodePayload(r, &payload); e != nil {
		return e.Code, nil, e
	}
	rsrc := route.Resource()
	changes, base := rsrc.Validator().Prepare(ctx, payload, nil, false)
	// Append lookup fields to base payload so it isn't caught by ReadOnly
	// (i.e.: contains id and parent resource refs if any).
	for k, v := range route.ResourcePath.Values() {
		base[k] = v
	}
	doc, errs := rsrc.Validator().Validate(changes, base)
	if len(errs) > 0 {
		return 422, nil, &Error{422, "Document contains error(s)", errs}
	}
	item, err := resource.NewItem(doc)
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	// TODO: add support for batch insert
	if err = rsrc.Insert(ctx, []*resource.Item{item}); err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	// Evaluate projection so response gets the same format as read requests.
	item.Payload, err = q.Projection.Eval(ctx, item.Payload, restResource{rsrc})
	if err != nil {
		e = NewError(err)
		return e.Code, nil, e
	}
	// See https://www.subbu.org/blog/2008/10/location-vs-content-location
	headers = http.Header{}
	itemID := item.ID
	if f := rsrc.Validator().GetField("id"); f != nil {
		if s, ok := f.Validator.(schema.FieldSerializer); ok {
			if tmp, err := s.Serialize(itemID); err == nil {
				itemID = tmp
			}
		}
	}
	headers.Set("Content-Location", fmt.Sprintf("%s/%s", r.URL.Path, itemID))
	return 201, headers, item
}
