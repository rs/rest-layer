package rest

import (
	"context"
	md5 "crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/rest-layer/resource"
)

// ResponseFormatter defines an interface responsible for formatting a the
// different types of response objects.
type ResponseFormatter interface {
	// FormatItem formats a single item in a format ready to be serialized by the ResponseSender
	FormatItem(ctx context.Context, headers http.Header, i *resource.Item, skipBody bool) (context.Context, interface{})
	// FormatList formats a list of items in a format ready to be serialized by the ResponseSender
	FormatList(ctx context.Context, headers http.Header, l *resource.ItemList, skipBody bool) (context.Context, interface{})
	// FormatError formats a REST formated error or a simple error in a format ready to be serialized by the ResponseSender
	FormatError(ctx context.Context, headers http.Header, err error, skipBody bool) (context.Context, interface{})
}

// ResponseSender defines an interface responsible for serializing and sending
// the response to the http.ResponseWriter.
type ResponseSender interface {
	// Send serialize the body, sets the given headers and write everything to
	// the provided response writer.
	Send(ctx context.Context, w http.ResponseWriter, status int, headers http.Header, body interface{})
}

// DefaultResponseFormatter provides a base response formatter to be used by
// default. This formatter can easily be extended or replaced by implementing
// ResponseFormatter interface and setting it on Handler.ResponseFormatter.
type DefaultResponseFormatter struct {
}

// DefaultResponseSender provides a base response sender to be used by default.
// This sender can easily be extended or replaced by implementing ResponseSender
// interface and setting it on Handler.ResponseSender.
type DefaultResponseSender struct {
}

// Send sends headers with the given status and marshal the data in JSON.
func (s DefaultResponseSender) Send(ctx context.Context, w http.ResponseWriter, status int, headers http.Header, body interface{}) {
	headers.Set("Content-Type", "application/json")
	// Apply headers to the response
	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(status)

	if body != nil {
		j, err := json.Marshal(body)
		if err != nil {
			w.WriteHeader(500)
			logErrorf(ctx, "Can't build response: %v", err)
			msg := fmt.Sprintf("Can't build response: %q", err.Error())
			w.Write([]byte(fmt.Sprintf("{\"code\": 500, \"msg\": \"%s\"}", msg)))
			return
		}
		if _, err = w.Write(j); err != nil {
			logErrorf(ctx, "Can't send response: %v", err)
		}
	}
}

// FormatItem implements ResponseFormatter.
func (f DefaultResponseFormatter) FormatItem(ctx context.Context, headers http.Header, i *resource.Item, skipBody bool) (context.Context, interface{}) {
	if i.ETag != "" {
		headers.Set("Etag", `W/"`+i.ETag+`"`)
	}
	if !i.Updated.IsZero() {
		headers.Set("Last-Modified", i.Updated.In(time.UTC).Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	}
	if skipBody || i.Payload == nil {
		return ctx, nil
	}
	return ctx, i.Payload
}

// FormatList implements ResponseFormatter.
func (f DefaultResponseFormatter) FormatList(ctx context.Context, headers http.Header, l *resource.ItemList, skipBody bool) (context.Context, interface{}) {
	if l.Total >= 0 {
		headers.Set("X-Total", strconv.Itoa(l.Total))
	}
	if l.Offset > 0 {
		headers.Set("X-Offset", strconv.Itoa(l.Offset))
	}

	hash := md5.New()
	for _, item := range l.Items {
		if item.ETag != "" {
			hash.Write([]byte(item.ETag))
		}
	}
	headers.Set("ETag", `W/"`+fmt.Sprintf("%x", hash.Sum(nil))+`"`)

	if !skipBody {
		payload := make([]map[string]interface{}, len(l.Items))
		for i, item := range l.Items {
			// Clone item payload to add the etag to the items in the list.
			d := map[string]interface{}{}
			for k, v := range item.Payload {
				d[k] = v
			}
			if item.ETag != "" {
				d["_etag"] = item.ETag
			}
			payload[i] = d
		}
		return ctx, payload
	}
	return ctx, nil
}

// FormatError implements ResponseFormatter.
func (f DefaultResponseFormatter) FormatError(ctx context.Context, headers http.Header, err error, skipBody bool) (context.Context, interface{}) {
	code := 500
	message := "Server Error"
	if err != nil {
		message = err.Error()
		if e, ok := err.(*Error); ok {
			code = e.Code
		}
	}
	if code >= 500 {
		logErrorf(ctx, "Server error: %v", err)
	}
	if !skipBody {
		payload := map[string]interface{}{
			"code":    code,
			"message": message,
		}
		if e, ok := err.(*Error); ok {
			if e.Issues != nil {
				payload["issues"] = e.Issues
			}
		}
		return ctx, payload
	}
	return ctx, nil
}

// formatResponse routes the type of response on the right ResponseFormater method for
// internally supported types.
func formatResponse(ctx context.Context, f ResponseFormatter, w http.ResponseWriter, status int, headers http.Header, resp interface{}, skipBody bool) (context.Context, int, interface{}) {
	var body interface{}
	switch resp := resp.(type) {
	case *resource.Item:
		ctx, body = f.FormatItem(ctx, headers, resp, skipBody)
	case *resource.ItemList:
		ctx, body = f.FormatList(ctx, headers, resp, skipBody)
	case *Error:
		if status == 0 {
			status = resp.Code
		}
		ctx, body = f.FormatError(ctx, headers, resp, skipBody)
	case error:
		if status == 0 {
			status = 500
		}
		ctx, body = f.FormatError(ctx, headers, resp, skipBody)
	default:
		// Let the response sender handle all other types of responses.
		// Even if the default response sender doesn't know how to handle
		// a type, nothing prevents a custom response sender from handling it.
		body = resp
	}
	return ctx, status, body
}
