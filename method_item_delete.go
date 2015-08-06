package rest

import "golang.org/x/net/context"

// itemDelete handles DELETE resquests on an item URL
func (r *request) itemDelete(ctx context.Context, route route) {
	lookup, err := route.lookup()
	if err != nil {
		r.sendError(err)
	}
	l, err := route.resource.handler.Find(ctx, lookup, 1, 1)
	if err != nil {
		r.sendError(err)
		return
	}
	if len(l.Items) == 0 {
		r.sendError(NotFoundError)
		return
	}
	original := l.Items[0]
	// If-Match / If-Unmodified-Since handling
	if err := r.checkIntegrityRequest(original); err != nil {
		r.sendError(err)
		return
	}
	if err := route.resource.handler.Delete(ctx, original); err != nil {
		r.sendError(err)
	} else {
		r.send(204, map[string]interface{}{})
	}
}
