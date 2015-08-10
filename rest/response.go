package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/rest-layer/resource"
	"golang.org/x/net/context"
)

// ResponseSender defines an interface responsible for formating, serializing and sending the response
// to the http.ResponseWriter.
type ResponseSender interface {
	// Send serialize the body, sets the given headers and write everything to the provided response writer
	Send(ctx context.Context, w http.ResponseWriter, status int, headers http.Header, body interface{})
	// SendError formats a REST formated error or a simple error in a format ready to be serialized by Send
	SendError(ctx context.Context, headers http.Header, err error, skipBody bool) (context.Context, interface{})
	// SendItem formats a single item in a format ready to be serialized by Send
	SendItem(ctx context.Context, headers http.Header, i *resource.Item, skipBody bool) (context.Context, interface{})
	// SendItem formats a list of items in a format ready to be serialized by Send
	SendList(ctx context.Context, headers http.Header, l *resource.ItemList, skipBody bool) (context.Context, interface{})
}

// DefaultResponseSender provides a base response sender to be used by default. This sender can
// easily be extended or replaced by implementing ResponseSender interface and setting it on Handler.ResponseSender.
type DefaultResponseSender struct {
}

// Send sends headers with the given status and marshal the data in JSON
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
			log.Printf("Can't build response: %s", err)
			msg := fmt.Sprintf("Can't build response: %s", strconv.Quote(err.Error()))
			w.Write([]byte(fmt.Sprintf("{\"code\": 500, \"msg\": \"%s\"}", msg)))
			return
		}
		w.Write(j)
	}
}

// SendError writes a REST formated error on the http.ResponseWriter
func (s DefaultResponseSender) SendError(ctx context.Context, headers http.Header, err error, skipBody bool) (context.Context, interface{}) {
	code := 500
	message := "Server Error"
	if err != nil {
		message = err.Error()
		if e, ok := err.(*Error); ok {
			code = e.Code
		}
	}
	if code >= 500 {
		log.Print(err.Error())
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

// SendItem sends a single item REST response on http.ResponseWriter
func (s DefaultResponseSender) SendItem(ctx context.Context, headers http.Header, i *resource.Item, skipBody bool) (context.Context, interface{}) {
	if i.ETag != "" {
		headers.Set("Etag", i.ETag)
	}
	if !i.Updated.IsZero() {
		headers.Set("Last-Modified", i.Updated.In(time.UTC).Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	}
	if skipBody {
		return ctx, nil
	}
	return ctx, i.Payload
}

// SendList sends a list of items as REST response on http.ResponseWriter
func (s DefaultResponseSender) SendList(ctx context.Context, headers http.Header, l *resource.ItemList, skipBody bool) (context.Context, interface{}) {
	if l.Total >= 0 {
		headers.Set("X-Total", strconv.FormatInt(int64(l.Total), 10))
	}
	if l.Page > 0 {
		headers.Set("X-Page", strconv.FormatInt(int64(l.Page), 10))
	}
	if !skipBody {
		payload := make([]map[string]interface{}, len(l.Items))
		for i, item := range l.Items {
			// Clone item payload to add the etag to the items in the list
			d := map[string]interface{}{}
			for k, v := range item.Payload {
				d[k] = v
			}
			d["_etag"] = item.ETag
			payload[i] = d
		}
		return ctx, payload
	}
	return ctx, nil
}
