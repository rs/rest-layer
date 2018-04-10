package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/rest-layer/resource"
)

// getMethodHandler returns the method handler for a given HTTP method in item
// or resource mode.
func getMethodHandler(isItem bool, method string) methodHandler {
	if isItem {
		switch method {
		case http.MethodOptions:
			return itemOptions
		case http.MethodHead, http.MethodGet:
			return itemGet
		case http.MethodPut:
			return itemPut
		case http.MethodPatch:
			return itemPatch
		case http.MethodDelete:
			return itemDelete
		}
	} else {
		switch method {
		case http.MethodOptions:
			return listOptions
		case http.MethodHead, http.MethodGet:
			return listGet
		case http.MethodPost:
			return listPost
		case http.MethodDelete:
			return listDelete
		}
	}
	return nil
}

// isMethodAllowed returns true if the method is allowed by the configuration.
func isMethodAllowed(isItem bool, method string, conf resource.Conf) bool {
	if isItem {
		switch method {
		case http.MethodOptions:
			return true
		case http.MethodHead, http.MethodGet:
			return conf.IsModeAllowed(resource.Read)
		case http.MethodPut:
			return conf.IsModeAllowed(resource.Create) || conf.IsModeAllowed(resource.Replace)
		case http.MethodPatch:
			return conf.IsModeAllowed(resource.Update)
		case http.MethodDelete:
			return conf.IsModeAllowed(resource.Delete)
		}
	} else {
		switch method {
		case http.MethodOptions:
			return true
		case http.MethodHead, http.MethodGet:
			return conf.IsModeAllowed(resource.List)
		case http.MethodPost:
			return conf.IsModeAllowed(resource.Create)
		case http.MethodDelete:
			return conf.IsModeAllowed(resource.Clear)
		}
	}
	return false
}

// getAllowedMethodHandler returns the method handler for the requested method
// if the resource configuration allows it.
func getAllowedMethodHandler(isItem bool, method string, conf resource.Conf) methodHandler {
	if isMethodAllowed(isItem, method, conf) {
		return getMethodHandler(isItem, method)
	}
	return nil
}

// setAllowHeader builds a Allow header based on the resource configuration.
func setAllowHeader(headers http.Header, isItem bool, conf resource.Conf) {
	methods := []string{}
	if isItem {
		// Methods are sorted
		if conf.IsModeAllowed(resource.Update) {
			methods = append(methods, "DELETE")
		}
		if conf.IsModeAllowed(resource.Read) {
			methods = append(methods, "GET, HEAD")
		}
		if conf.IsModeAllowed(resource.Update) {
			methods = append(methods, "PATCH")
			// See http://tools.ietf.org/html/rfc5789#section-3
			headers.Set("Allow-Patch", "application/json")
		}
		if conf.IsModeAllowed(resource.Create) || conf.IsModeAllowed(resource.Replace) {
			methods = append(methods, "PUT")
		}
	} else {
		// Methods are sorted
		if conf.IsModeAllowed(resource.Clear) {
			methods = append(methods, "DELETE")
		}
		if conf.IsModeAllowed(resource.List) {
			methods = append(methods, "GET, HEAD")
		}
		if conf.IsModeAllowed(resource.Create) {
			methods = append(methods, "POST")
		}
	}
	if len(methods) > 0 {
		headers.Set("Allow", strings.Join(methods, ", "))
	}
}

// compareEtag compares a client provided etag with a base etag. The client
// provided etag may or may not have quotes while the base etag is never quoted.
// This loose comparison of etag allows clients not strictly respecting RFC to
// send the etag with or without quotes when the etag comes from, for instance,
// the API JSON response.
func compareEtag(etag, baseEtag string) bool {
	if etag == "" {
		return false
	}
	if strings.HasPrefix(etag, "W/") {
		if etag[2:] == baseEtag {
			return true
		}
		if l := len(etag); l == len(baseEtag)+4 && l > 4 && etag[2] == '"' && etag[l-1] == '"' && etag[3:l-1] == baseEtag {
			return true
		}
	}
	return false
}

// decodePayload decodes the payload from the provided request.
func decodePayload(r *http.Request, payload *map[string]interface{}) *Error {
	// Check content-type, if not specified, assume it's JSON and fail later
	if ct := r.Header.Get("Content-Type"); ct != "" && strings.TrimSpace(strings.SplitN(ct, ";", 2)[0]) != "application/json" {
		return &Error{501, fmt.Sprintf("Invalid Content-Type header: `%s' not supported", ct), nil}
	}
	if r.Body == nil {
		return nil
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(payload); err != nil {
		return &Error{400, fmt.Sprintf("Malformed body: %v", err), nil}
	}
	return nil
}

// checkIntegrityRequest ensures that original item exists and complies with
// conditions expressed by If-Match and/or If-Unmodified-Since headers if
// present.
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
			} else if original.Updated.Truncate(time.Second).After(ifUnmodTime) {
				// Item's update time is truncated to the second because RFC1123 doesn't support more
				return ErrPreconditionFailed
			}
		}
	}
	return nil
}

func logErrorf(ctx context.Context, format string, a ...interface{}) {
	if resource.Logger != nil {
		resource.Logger(ctx, resource.LogLevelError, fmt.Sprintf(format, a...), nil)
	}
}
