package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/rest-layer/schema"
	"golang.org/x/net/context"
)

// requestHandler handles the request life cycle
type requestHandler struct {
	ctx      context.Context
	root     *RootResource
	req      *http.Request
	res      http.ResponseWriter
	s        ResponseSender
	skipBody bool
}

func (r *requestHandler) send(status int, data interface{}) {
	r.s.Send(r.res, status, data)
}
func (r *requestHandler) sendError(err error) {
	r.s.SendError(r.res, err, r.skipBody)
}
func (r *requestHandler) sendItem(status int, i *Item) {
	r.s.SendItem(r.res, status, i, r.skipBody)
}
func (r *requestHandler) sendList(l *ItemList) {
	r.s.SendList(r.res, l, r.skipBody)
}

// decodePayload decodes the payload from the provided request
func (r *requestHandler) decodePayload(payload *map[string]interface{}) *Error {
	// Check content-type, if not specified, assume it's JSON and fail later
	if ct := r.req.Header.Get("Content-Type"); ct != "" && ct != "application/json" {
		return &Error{501, fmt.Sprintf("Invalid Content-Type header: `%s' not supported", ct), nil}
	}
	decoder := json.NewDecoder(r.req.Body)
	defer r.req.Body.Close()
	if err := decoder.Decode(payload); err != nil {
		return &Error{400, fmt.Sprintf("Malformed body: %s", err.Error()), nil}
	}
	return nil
}

// checkIntegrityRequest ensures that orignal item exists and complies with conditions
// expressed by If-Match and/or If-Unmodified-Since headers if present.
func (r *requestHandler) checkIntegrityRequest(original *Item) *Error {
	ifMatch := r.req.Header.Get("If-Match")
	ifUnmod := r.req.Header.Get("If-Unmodified-Since")
	if ifMatch != "" || ifUnmod != "" {
		if original == nil {
			return NotFoundError
		}
		if ifMatch != "" && original.Etag != ifMatch {
			return PreconditionFailedError
		}
		if ifUnmod != "" {
			if ifUnmodTime, err := time.Parse(time.RFC1123, ifUnmod); err != nil {
				return &Error{400, "Invalid If-Unmodified-Since header", nil}
			} else if original.Updated.After(ifUnmodTime) {
				return PreconditionFailedError
			}
		}
	}
	return nil
}

// checkReferences checks that fields with the Reference validator reference an existing object
func (r *requestHandler) checkReferences(payload map[string]interface{}, s schema.Validator) *Error {
	for name, value := range payload {
		field := s.GetField(name)
		if field == nil {
			continue
		}
		// Check reference if validator is of type Reference
		if field.Validator != nil {
			if ref, ok := field.Validator.(*schema.Reference); ok {
				resource := r.root.GetResource(ref.Path)
				if resource == nil {
					return &Error{500, fmt.Sprintf("Invalid resource reference for field `%s': %s", name, ref.Path), nil}
				}
				lookup := NewLookup()
				lookup.Fields["id"] = value
				list, _ := resource.handler.Find(lookup, 1, 1, r.ctx)
				if len(list.Items) == 0 {
					return &Error{404, fmt.Sprintf("Resource reference not found for field `%s'", name), nil}
				}
			}
		}
		// Check sub-schema if any
		if field.Schema != nil && value != nil {
			if subPayload, ok := value.(map[string]interface{}); ok {
				if err := r.checkReferences(subPayload, field.Schema); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
