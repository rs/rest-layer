package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// ResponseSender defines an interface responsible for serializing and sending the response
// to the http.ResponseWriter.
type ResponseSender interface {
	// Send sends headers with the given status and marshal the data in JSON
	Send(w http.ResponseWriter, status int, data interface{})
	// SendError writes a REST formated error on the http.ResponseWriter
	SendError(w http.ResponseWriter, err error, skipBody bool)
	// SendItem sends a single item REST response on http.ResponseWriter
	SendItem(w http.ResponseWriter, status int, i *Item, skipBody bool)
	// SendList sends a list of items as REST response on http.ResponseWriter.
	SendList(w http.ResponseWriter, l *ItemList, skipBody bool)
}

// DefaultResponseSender provides a base response sender to be used by default. This sender can
// easily be extended or replaced.
type DefaultResponseSender struct {
}

// Send sends headers with the given status and marshal the data in JSON
func (s DefaultResponseSender) Send(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		j, err := json.Marshal(data)
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
func (s DefaultResponseSender) SendError(w http.ResponseWriter, err error, skipBody bool) {
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
			payload["issues"] = e.Issues
		}
		s.Send(w, code, payload)
	} else {
		s.Send(w, code, nil)
	}
}

// SendItem sends a single item REST response on http.ResponseWriter
func (s DefaultResponseSender) SendItem(w http.ResponseWriter, status int, i *Item, skipBody bool) {
	if i.Etag != "" {
		w.Header().Set("Etag", i.Etag)
	}
	if !i.Updated.IsZero() {
		w.Header().Set("Last-Modified", i.Updated.In(time.UTC).Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	}
	if skipBody {
		s.Send(w, status, nil)
	} else {
		s.Send(w, status, i.Payload)
	}
}

// SendList sends a list of items as REST response on http.ResponseWriter
func (s DefaultResponseSender) SendList(w http.ResponseWriter, l *ItemList, skipBody bool) {
	if l.Total >= 0 {
		w.Header().Set("X-Total", strconv.FormatInt(int64(l.Total), 10))
	}
	if l.Page > 0 {
		w.Header().Set("X-Page", strconv.FormatInt(int64(l.Page), 10))
	}
	if !skipBody {
		payload := make([]map[string]interface{}, len(l.Items))
		for i, item := range l.Items {
			// Clone item payload to add the etag to the items in the list
			d := map[string]interface{}{}
			for k, v := range item.Payload {
				d[k] = v
			}
			d["_etag"] = item.Etag
			payload[i] = d
		}
		s.Send(w, 200, payload)
	} else {
		s.Send(w, 200, nil)

	}
}
