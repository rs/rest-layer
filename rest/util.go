package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
)

// compareEtag compares a client provided etag with a base etag. The client provided
// etag may or may not have quotes while the base etag is never quoted. This loose
// comparison of etag allows clients not stricly respecting RFC to send the etag with
// or without quotes when the etag comes from, for instance, the API JSON response.
func compareEtag(etag, baseEtag string) bool {
	if etag == baseEtag {
		return true
	}
	if l := len(etag); l == len(baseEtag)+2 && l > 3 && etag[0] == '"' && etag[l-1] == '"' && etag[1:l-1] == baseEtag {
		return true
	}
	return false
}

// decodePayload decodes the payload from the provided request
func decodePayload(r *http.Request, payload *map[string]interface{}) *Error {
	// Check content-type, if not specified, assume it's JSON and fail later
	if ct := r.Header.Get("Content-Type"); ct != "" && strings.TrimSpace(strings.SplitN(ct, ";", 2)[0]) != "application/json" {
		return &Error{501, fmt.Sprintf("Invalid Content-Type header: `%s' not supported", ct), nil}
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(payload); err != nil {
		return &Error{400, fmt.Sprintf("Malformed body: %s", err.Error()), nil}
	}
	return nil
}

// checkIntegrityRequest ensures that orignal item exists and complies with conditions
// expressed by If-Match and/or If-Unmodified-Since headers if present.
func checkIntegrityRequest(r *http.Request, original *resource.Item) *Error {
	ifMatch := r.Header.Get("If-Match")
	ifUnmod := r.Header.Get("If-Unmodified-Since")
	if ifMatch != "" || ifUnmod != "" {
		if original == nil {
			return ErrNotFound
		}
		if ifMatch != "" && !compareEtag(ifMatch, original.ETag) {
			return ErrPreconditionFailed
		}
		if ifUnmod != "" {
			if ifUnmodTime, err := time.Parse(time.RFC1123, ifUnmod); err != nil {
				return &Error{400, "Invalid If-Unmodified-Since header", nil}
			} else if original.Updated.After(ifUnmodTime) {
				return ErrPreconditionFailed
			}
		}
	}
	return nil
}

// checkReferences ensures that fields with the Reference validator reference an existing object
func checkReferences(ctx context.Context, payload map[string]interface{}, s schema.Validator) *Error {
	for name, value := range payload {
		field := s.GetField(name)
		if field == nil {
			continue
		}
		// Check reference if validator is of type Reference
		if field.Validator != nil {
			if ref, ok := field.Validator.(*schema.Reference); ok {
				router, ok := IndexFromContext(ctx)
				if !ok {
					return &Error{500, "Router not available in context", nil}
				}
				rsrc, _, found := router.GetResource(ref.Path)
				if !found {
					return &Error{500, fmt.Sprintf("Invalid resource reference for field `%s': %s", name, ref.Path), nil}
				}
				l := resource.NewLookup()
				l.AddQuery(schema.Query{schema.Equal{Field: "id", Value: value}})
				list, _ := rsrc.Find(ctx, l, 1, 1)
				if len(list.Items) == 0 {
					return &Error{404, fmt.Sprintf("Resource reference not found for field `%s'", name), nil}
				}
			}
		}
		// Check sub-schema if any
		if field.Schema != nil && value != nil {
			if subPayload, ok := value.(map[string]interface{}); ok {
				if err := checkReferences(ctx, subPayload, field.Schema); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
