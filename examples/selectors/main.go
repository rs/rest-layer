package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"
)

/* Play with it:
http :8080/posts fields=='id,thumbnail_url(height=80):thumb_s_url'
http :8080/posts fields=='id:i,meta{title:t, body:b}:m,thumbnail_url(height=80):thumb_small_url'
*/

var (
	post = schema.Schema{
		"id":      schema.IDField,
		"created": schema.CreatedField,
		"updated": schema.UpdatedField,
		"thumbnail_url": schema.Field{
			Params: &schema.Params{
				// Appends a "w" and/or "h" query string parameter(s) to the value (URL) if width or height params passed
				Handler: func(value interface{}, params map[string]interface{}) (interface{}, error) {
					str, ok := value.(string)
					if !ok {
						return nil, errors.New("not a string")
					}
					sep := "?"
					if strings.IndexByte(str, '?') > 0 {
						sep = "&"
					}
					if width, found := params["width"]; found {
						str = fmt.Sprintf("%s%sw=%d", str, sep, width)
						sep = "&"
					}
					if height, found := params["height"]; found {
						str = fmt.Sprintf("%s%sy=%d", str, sep, height)
					}
					return str, nil
				},
				Validators: map[string]schema.FieldValidator{
					"width": schema.Integer{
						Boundaries: &schema.Boundaries{Max: 1000},
					},
					"height": schema.Integer{
						Boundaries: &schema.Boundaries{Max: 1000},
					},
				},
			},
		},
		"meta": schema.Field{
			Schema: &schema.Schema{
				"title": schema.Field{},
				"body":  schema.Field{},
			},
		},
	}
)

func main() {
	index := resource.NewIndex()

	index.Bind("posts", resource.New(post, mem.NewHandler(), resource.Conf{
		// Posts can only be read, created and deleted, not updated
		AllowedModes: resource.ReadWrite,
	}))

	// Create API HTTP handler for the resource graph
	api, err := rest.NewHandler(index)
	if err != nil {
		log.Fatalf("Invalid API configuration: %s", err)
	}

	// Bind the API under /api/ path
	http.Handle("/", api)

	// Serve it
	log.Print("Serving API on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
