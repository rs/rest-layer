package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"golang.org/x/net/context"

	"github.com/justinas/alice"
	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/xaccess"
	"github.com/rs/xlog"
)

var (
	user = schema.Schema{
		Fields: schema.Fields{
			"id": {
				Required:   true,
				ReadOnly:   true,
				Filterable: true,
				Sortable:   true,
				Validator: &schema.String{
					Regexp: "^[0-9a-z_-]{2,150}$",
				},
			},
			"created":  schema.CreatedField,
			"updated":  schema.UpdatedField,
			"name":     {},
			"ip":       {Validator: &schema.IP{StoreBinary: true}},
			"password": schema.PasswordField,
		},
	}

	postFollower = schema.Schema{
		Fields: schema.Fields{
			"id": schema.IDField,
			"post": {
				Validator: &schema.Reference{Path: "posts"},
			},
			"user": {
				Filterable: true,
				Sortable:   true,
				Validator:  &schema.Reference{Path: "users"},
			},
		},
	}

	post = schema.Schema{
		Fields: schema.Fields{
			"id":      schema.IDField,
			"created": schema.CreatedField,
			"updated": schema.UpdatedField,
			"user": {
				Validator: &schema.Reference{Path: "users"},
			},
			"thumbnail_url": {
				Params: schema.Params{
					"width": {
						Validator: schema.Integer{
							Boundaries: &schema.Boundaries{Max: 1000},
						},
					},
					"height": {
						Validator: schema.Integer{
							Boundaries: &schema.Boundaries{Max: 1000},
						},
					},
				},
				// Appends a "w" and/or "h" query string parameter(s) to the value (URL) if width or height params passed
				Handler: func(ctx context.Context, value interface{}, params map[string]interface{}) (interface{}, error) {
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
			},
			"meta": {
				Schema: &schema.Schema{
					Fields: schema.Fields{
						"title": {},
						"body":  {},
					},
				},
			},
		},
	}
)

func main() {
	index := resource.NewIndex()

	index.Bind("users", user, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
	})

	posts := index.Bind("posts", post, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
	})

	posts.Bind("followers", "post", postFollower, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
	})

	// Create API HTTP handler for the resource graph
	api, err := rest.NewHandler(index)
	if err != nil {
		log.Fatalf("Invalid API configuration: %s", err)
	}

	// Setup logger
	c := alice.New()
	c.Append(xlog.NewHandler(xlog.Config{}))
	c.Append(xaccess.NewHandler())
	resource.LoggerLevel = resource.LogLevelDebug
	resource.Logger = func(ctx context.Context, level resource.LogLevel, msg string, fields map[string]interface{}) {
		xlog.FromContext(ctx).OutputF(xlog.Level(level), 2, msg, fields)
	}

	// Bind the API under the root path
	http.Handle("/", c.Then(api))

	// Inject some fixtures
	fixtures := [][]string{
		{"PUT", "/users/johndoe", `{"name": "John Doe", "ip": "1.2.3.4", "password": "secret"}`},
		{"PUT", "/users/fan1", `{"name": "Fan 1", "ip": "1.2.3.4", "password": "secret"}}`},
		{"PUT", "/users/fan2", `{"name": "Fan 2", "ip": "1.2.3.4", "password": "secret"}}`},
		{"PUT", "/users/fan3", `{"name": "Fan 3", "ip": "1.2.3.4", "password": "secret"}}`},
		{"PUT", "/users/fan4", `{"name": "Fan 4", "ip": "1.2.3.4", "password": "secret"}}`},
		{"PUT", "/posts/ar5qrgukj5l7a6eq2ps0",
			`{
				"user": "johndoe",
				"thumbnail_url": "http://dom.com/image.png",
				"meta": {
					"title": "First Post",
					"body": "This is my first post"
				}
			}`},
		{"POST", "/posts/ar5qrgukj5l7a6eq2ps0/followers", `{"user": "fan1"}`},
		{"POST", "/posts/ar5qrgukj5l7a6eq2ps0/followers", `{"user": "fan2"}`},
		{"POST", "/posts/ar5qrgukj5l7a6eq2ps0/followers", `{"user": "fan3"}`},
	}
	for _, fixture := range fixtures {
		req, err := http.NewRequest(fixture[0], fixture[1], strings.NewReader(fixture[2]))
		if err != nil {
			log.Fatal(err)
		}
		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		if w.Code >= 400 {
			log.Fatalf("Error returned for `%s %s`: %v", fixture[0], fixture[1], w)
		}
	}

	// Serve it
	log.Print("Serving API on http://localhost:8080")
	log.Println("Play with (httpie):\n",
		"- http :8080/posts fields=='id,thumb_s_url:thumbnail_url(height:80)'\n",
		"- http :8080/posts fields=='i:id,m:meta{t:title,b:body},thumb_small_url:thumbnail_url(height:80)'\n",
		"- http :8080/posts fields=='id,meta,user{id,name}'\n",
		"- http :8080/posts/ar5qrgukj5l7a6eq2ps0/followers fields=='post{id,meta{title}},user{id,name}'\n",
		"- http :8080/posts/ar5qrgukj5l7a6eq2ps0 fields=='id,meta{title},followers(limit:2){user{id,name}}'")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
