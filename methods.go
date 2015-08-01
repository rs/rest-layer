package rest

import (
	"fmt"
	"strconv"
	"time"
)

// handleListRequestGET handles GET resquests on a resource URL
func (r *requestHandler) handleListRequestGET(lookup *Lookup, resource *Resource) {
	page := 1
	perPage := 0
	if !r.skipBody {
		if resource.conf.PaginationDefaultLimit > 0 {
			perPage = resource.conf.PaginationDefaultLimit
		} else {
			// Default value on non HEAD request for perPage is -1 (pagination disabled)
			perPage = -1
		}
		if p := r.req.URL.Query().Get("page"); p != "" {
			i, err := strconv.ParseUint(p, 10, 32)
			if err != nil {
				r.sendError(&Error{422, "Invalid `page` paramter", nil})
				return
			}
			page = int(i)
		}
		if l := r.req.URL.Query().Get("limit"); l != "" {
			i, err := strconv.ParseUint(l, 10, 32)
			if err != nil {
				r.sendError(&Error{422, "Invalid `limit` paramter", nil})
				return
			}
			perPage = int(i)
		}
		if perPage == -1 && page != 1 {
			r.sendError(&Error{422, "Cannot use `page' parameter with no `limit' paramter on a resource with no default pagination size", nil})
		}
		if sort := r.req.URL.Query().Get("sort"); sort != "" {
			if err := lookup.SetSort(sort, resource.schema); err != nil {
				r.sendError(&Error{422, "Invalid `sort` paramter", nil})
				return
			}
		}
	}
	if filter := r.req.URL.Query().Get("filter"); filter != "" {
		if err := lookup.SetFilter(filter, resource.schema); err != nil {
			r.sendError(&Error{422, "Invalid `filter` parameter", nil})
			return
		}
	}
	list, err := resource.handler.Find(lookup, page, perPage, r.ctx)
	if err != nil {
		r.sendError(err)
		return
	}
	r.sendList(list)
}

// handleListRequestPOST handles POST resquests on a resource URL
func (r *requestHandler) handleListRequestPOST(lookup *Lookup, resource *Resource) {
	var payload map[string]interface{}
	if err := r.decodePayload(&payload); err != nil {
		r.sendError(err)
		return
	}
	changes, base := resource.schema.Prepare(payload, nil, false)
	// Append lookup fields to base payload so it isn't caught by ReadOnly
	// (i.e.: contains id and parent resource refs if any)
	lookup.applyFields(base)
	doc, errs := resource.schema.Validate(changes, base)
	if len(errs) > 0 {
		r.sendError(&Error{422, "Document contains error(s)", errs})
		return
	}
	// Check that fields with the Reference validator reference an existing object
	if err := r.checkReferences(doc, resource.schema); err != nil {
		r.sendError(err)
		return
	}
	item, err := NewItem(doc)
	if err != nil {
		r.sendError(err)
		return
	}
	// TODO: add support for batch insert
	if err := resource.handler.Insert([]*Item{item}, r.ctx); err != nil {
		r.sendError(err)
		return
	}
	// See https://www.subbu.org/blog/2008/10/location-vs-content-location
	r.res.Header().Set("Content-Location", fmt.Sprintf("/%s/%s", r.req.URL.Path, item.ID))
	r.sendItem(201, item)
}

// handleListRequestDELETE handles DELETE resquests on a resource URL
func (r *requestHandler) handleListRequestDELETE(lookup *Lookup, resource *Resource) {
	if total, err := resource.handler.Clear(lookup, r.ctx); err != nil {
		r.sendError(err)
	} else {
		r.res.Header().Set("X-Total", strconv.FormatInt(int64(total), 10))
		r.send(204, map[string]interface{}{})
	}
}

// handleItemRequestGET handles GET and HEAD resquests on an item URL
func (r *requestHandler) handleItemRequestGET(lookup *Lookup, resource *Resource) {
	l, err := resource.handler.Find(lookup, 1, 1, r.ctx)
	if err != nil {
		r.sendError(err)
		return
	} else if len(l.Items) == 0 {
		r.sendError(NotFoundError)
		return
	}
	item := l.Items[0]
	// Handle conditional request: If-None-Match
	if r.req.Header.Get("If-None-Match") == item.Etag {
		r.send(304, nil)
		return
	}
	// Handle conditional request: If-Modified-Since
	if r.req.Header.Get("If-Modified-Since") != "" {
		if ifModTime, err := time.Parse(time.RFC1123, r.req.Header.Get("If-Modified-Since")); err != nil {
			r.sendError(&Error{400, "Invalid If-Modified-Since header", nil})
			return
		} else if item.Updated.Equal(ifModTime) || item.Updated.Before(ifModTime) {
			r.send(304, nil)
			return
		}
	}
	r.sendItem(200, item)
}

// handleItemRequestPUT handles PUT resquests on an item URL
//
// Reference: http://tools.ietf.org/html/rfc2616#section-9.6
func (r *requestHandler) handleItemRequestPUT(lookup *Lookup, resource *Resource) {
	var payload map[string]interface{}
	if err := r.decodePayload(&payload); err != nil {
		r.sendError(err)
		return
	}
	// Fetch original item if exist (PUT can be used to create a document with a manual id)
	var original *Item
	if l, err := resource.handler.Find(lookup, 1, 1, r.ctx); err != nil && err != NotFoundError {
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
	if !resource.conf.isModeAllowed(mode) {
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
		changes, base = resource.schema.Prepare(payload, nil, false)
		status = 201
	} else {
		// PUT used to replace an existing document
		changes, base = resource.schema.Prepare(payload, &original.Payload, true)
	}
	// Append lookup fields to base payload so it isn't caught by ReadOnly
	// (i.e.: contains id and parent resource refs if any)
	lookup.applyFields(base)
	doc, errs := resource.schema.Validate(changes, base)
	if len(errs) > 0 {
		r.sendError(&Error{422, "Document contains error(s)", errs})
		return
	}
	// Check that fields with the Reference validator reference an existing object
	if err := r.checkReferences(doc, resource.schema); err != nil {
		r.sendError(err)
		return
	}
	if original != nil {
		if id, found := doc["id"]; found && id != original.ID {
			r.sendError(&Error{422, "Cannot change document ID", nil})
			return
		}
	}
	item, err := NewItem(doc)
	if err != nil {
		r.sendError(err)
		return
	}
	// If we have an original item, pass it to the handler so we make sure
	// we are still replacing the same version of the object as handler is
	// supposed check the original etag before storing when an original object
	// is provided.
	if err := resource.handler.Update(item, original, r.ctx); err != nil {
		r.sendError(err)
	} else {
		r.sendItem(status, item)
	}
}

// handleItemRequestPATCH handles PATCH resquests on an item URL
//
// Reference: http://tools.ietf.org/html/rfc5789
func (r *requestHandler) handleItemRequestPATCH(lookup *Lookup, resource *Resource) {
	var payload map[string]interface{}
	if err := r.decodePayload(&payload); err != nil {
		r.sendError(err)
		return
	}
	// Get original item if any
	var original *Item
	if l, err := resource.handler.Find(lookup, 1, 1, r.ctx); err != nil {
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
	changes, base := resource.schema.Prepare(payload, &original.Payload, false)
	// Append lookup fields to base payload so it isn't caught by ReadOnly
	// (i.e.: contains id and parent resource refs if any)
	lookup.applyFields(base)
	doc, errs := resource.schema.Validate(changes, base)
	if len(errs) > 0 {
		r.sendError(&Error{422, "Document contains error(s)", errs})
		return
	}
	// Check that fields with the Reference validator reference an existing object
	if err := r.checkReferences(doc, resource.schema); err != nil {
		r.sendError(err)
		return
	}
	item, err := NewItem(doc)
	if err != nil {
		r.sendError(err)
		return
	}
	// Store the modified document by providing the orignal doc to instruct
	// handler to ensure the stored document didn't change between in the
	// interval. An PreconditionFailedError will be thrown in case of race condition
	// (i.e.: another thread modified the document between the Find() and the Store())
	if err := resource.handler.Update(item, original, r.ctx); err != nil {
		r.sendError(err)
	} else {
		r.sendItem(200, item)
	}
}

// handleItemRequestDELETE handles DELETE resquests on an item URL
func (r *requestHandler) handleItemRequestDELETE(lookup *Lookup, resource *Resource) {
	l, err := resource.handler.Find(lookup, 1, 1, r.ctx)
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
	if err := resource.handler.Delete(original, r.ctx); err != nil {
		r.sendError(err)
	} else {
		r.send(204, map[string]interface{}{})
	}
}
