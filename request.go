package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/rest-layer/schema"
	"golang.org/x/net/context"
)

type requestHandler interface {
	itemGet(ctx context.Context, route RouteMatch) (int, http.Header, interface{})
	itemPut(ctx context.Context, route RouteMatch) (int, http.Header, interface{})
	itemPatch(ctx context.Context, route RouteMatch) (int, http.Header, interface{})
	itemDelete(ctx context.Context, route RouteMatch) (int, http.Header, interface{})
	listGet(ctx context.Context, route RouteMatch) (int, http.Header, interface{})
	listPost(ctx context.Context, route RouteMatch) (int, http.Header, interface{})
	listDelete(ctx context.Context, route RouteMatch) (int, http.Header, interface{})
}

// request handles the request life cycle
type request struct {
	req *http.Request
}

// decodePayload decodes the payload from the provided request
func (r *request) decodePayload(payload *map[string]interface{}) *Error {
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
func (r *request) checkIntegrityRequest(original *Item) *Error {
	ifMatch := r.req.Header.Get("If-Match")
	ifUnmod := r.req.Header.Get("If-Unmodified-Since")
	if ifMatch != "" || ifUnmod != "" {
		if original == nil {
			return NotFoundError
		}
		if ifMatch != "" && original.ETag != ifMatch {
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
func (r *request) checkReferences(ctx context.Context, payload map[string]interface{}, s schema.Validator) *Error {
	for name, value := range payload {
		field := s.GetField(name)
		if field == nil {
			continue
		}
		// Check reference if validator is of type Reference
		if field.Validator != nil {
			if ref, ok := field.Validator.(*schema.Reference); ok {
				router, ok := RouterFromContext(ctx)
				if !ok {
					return &Error{500, "Router not available in context", nil}
				}
				resource, _, found := router.GetResource(ref.Path)
				if !found {
					return &Error{500, fmt.Sprintf("Invalid resource reference for field `%s': %s", name, ref.Path), nil}
				}
				l := lookup{filter: schema.Query{schema.Equal{Field: "id", Value: value}}}
				list, _ := resource.handler.Find(ctx, &l, 1, 1)
				if len(list.Items) == 0 {
					return &Error{404, fmt.Sprintf("Resource reference not found for field `%s'", name), nil}
				}
			}
		}
		// Check sub-schema if any
		if field.Schema != nil && value != nil {
			if subPayload, ok := value.(map[string]interface{}); ok {
				if err := r.checkReferences(ctx, subPayload, field.Schema); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
