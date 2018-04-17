package graphql

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/rs/rest-layer/resource/testing/mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

var (
	user = schema.Schema{
		Description: "Defines user information",
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
			"created": schema.CreatedField,
			"updated": schema.UpdatedField,
			"name":    {},
			"admin": {
				Filterable: true,
				Validator:  &schema.Bool{},
			},
			"ip":       {Validator: &schema.IP{StoreBinary: true}},
			"password": schema.PasswordField,
		},
	}

	postFollower = schema.Schema{
		Description: "Link a post to its followers",
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
		Description: "Defines a blog post",
		Fields: schema.Fields{
			"id":      schema.IDField,
			"created": schema.CreatedField,
			"updated": schema.UpdatedField,
			"user": {
				Validator: &schema.Reference{Path: "users"},
			},
			"thumbnail_url": {
				Description: "Resizable thumbnail URL for a post. Use width and height parameters to get a specific size.",
				Params: schema.Params{
					"width": {
						Description: "Change the width of the thumbnail to the value in pixels",
						Validator: schema.Integer{
							Boundaries: &schema.Boundaries{Max: 1000},
						},
					},
					"height": {
						Description: "Change the height of the thumbnail to the value in pixels",
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

func performRequest(h http.Handler, r *http.Request) (int, string) {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	body, _ := ioutil.ReadAll(w.Body)
	return w.Code, string(body)
}

func TestHandler(t *testing.T) {
	oldLogger := resource.Logger
	resource.Logger = nil
	defer func() { resource.Logger = oldLogger }()
	index := resource.NewIndex()

	users := index.Bind("users", user, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
	})

	users.Alias("admin", url.Values{"filter": []string{`{"admin": true}`}})

	posts := index.Bind("posts", post, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
	})

	posts.Bind("followers", "post", postFollower, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
	})

	// Inject some fixtures
	fixtures := [][]string{
		{"PUT", "/users/johndoe", `{"name": "John Doe", "ip": "1.2.3.4", "password": "secret", "admin": true}`},
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
	api, err := rest.NewHandler(index)
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

	gql, err := NewHandler(index)
	assert.NoError(t, err)

	r, _ := http.NewRequest("GET", "/?query={postsList{id,thumb_s_url:thumbnail_url(height:80)}}", nil)
	s, b := performRequest(gql, r)
	assert.Equal(t, 200, s)
	assert.Equal(t, "{\"data\":{\"postsList\":[{\"id\":\"ar5qrgukj5l7a6eq2ps0\",\"thumb_s_url\":\"http://dom.com/image.png?y=80\"}]}}\n", b)
	r, _ = http.NewRequest("GET", "/?query={postsList{i:id,m:meta{t:title, b:body},thumb_small_url:thumbnail_url(height:80)}}", nil)
	s, b = performRequest(gql, r)
	assert.Equal(t, 200, s)
	assert.Equal(t, "{\"data\":{\"postsList\":[{\"i\":\"ar5qrgukj5l7a6eq2ps0\",\"m\":{\"b\":\"This is my first post\",\"t\":\"First Post\"},\"thumb_small_url\":\"http://dom.com/image.png?y=80\"}]}}\n", b)
	r, _ = http.NewRequest("GET", "/?query={postsList{id,meta{title},user{id,name}}}", nil)
	s, b = performRequest(gql, r)
	assert.Equal(t, 200, s)
	assert.Equal(t, "{\"data\":{\"postsList\":[{\"id\":\"ar5qrgukj5l7a6eq2ps0\",\"meta\":{\"title\":\"First Post\"},\"user\":{\"id\":\"johndoe\",\"name\":\"John Doe\"}}]}}\n", b)
	r, _ = http.NewRequest("GET", "/?query={posts(id:\"ar5qrgukj5l7a6eq2ps0\"){followers{post{id,meta{title}},user{id,name}}}}", nil)
	s, b = performRequest(gql, r)
	assert.Equal(t, 200, s)
	assert.Equal(t, "{\"data\":{\"posts\":{\"followers\":[{\"post\":{\"id\":\"ar5qrgukj5l7a6eq2ps0\",\"meta\":{\"title\":\"First Post\"}},\"user\":{\"id\":\"fan1\",\"name\":\"Fan 1\"}},{\"post\":{\"id\":\"ar5qrgukj5l7a6eq2ps0\",\"meta\":{\"title\":\"First Post\"}},\"user\":{\"id\":\"fan2\",\"name\":\"Fan 2\"}},{\"post\":{\"id\":\"ar5qrgukj5l7a6eq2ps0\",\"meta\":{\"title\":\"First Post\"}},\"user\":{\"id\":\"fan3\",\"name\":\"Fan 3\"}}]}}}\n", b)
	r, _ = http.NewRequest("GET", "/?query={posts(id:\"ar5qrgukj5l7a6eq2ps0\"){id,meta{title},followers(limit:2){user{id,name}}}}", nil)
	s, b = performRequest(gql, r)
	assert.Equal(t, 200, s)
	assert.Equal(t, "{\"data\":{\"posts\":{\"followers\":[{\"user\":{\"id\":\"fan1\",\"name\":\"Fan 1\"}},{\"user\":{\"id\":\"fan2\",\"name\":\"Fan 2\"}}],\"id\":\"ar5qrgukj5l7a6eq2ps0\",\"meta\":{\"title\":\"First Post\"}}}}\n", b)

	r, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{postsList{id,thumb_s_url:thumbnail_url(height:80)}}"))
	s, b = performRequest(gql, r)
	assert.Equal(t, 200, s)
	assert.Equal(t, "{\"data\":{\"postsList\":[{\"id\":\"ar5qrgukj5l7a6eq2ps0\",\"thumb_s_url\":\"http://dom.com/image.png?y=80\"}]}}\n", b)

	r, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"query\":\"{postsList{id,thumb_s_url:thumbnail_url(height:80)}}\"}"))
	r.Header.Set("Content-Type", "application/json")
	s, b = performRequest(gql, r)
	assert.Equal(t, 200, s)
	assert.Equal(t, "{\"data\":{\"postsList\":[{\"id\":\"ar5qrgukj5l7a6eq2ps0\",\"thumb_s_url\":\"http://dom.com/image.png?y=80\"}]}}\n", b)

	r, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{invalid json"))
	r.Header.Set("Content-Type", "application/json")
	s, b = performRequest(gql, r)
	assert.Equal(t, 400, s)
	assert.Equal(t, "Cannot unmarshal JSON: invalid character 'i' looking for beginning of object key string\n{\"data\":null,\"errors\":[{\"message\":\"Must provide an operation.\",\"locations\":[]}]}\n", b)

	r, _ = http.NewRequest("PUT", "/", nil)
	s, b = performRequest(gql, r)
	assert.Equal(t, 405, s)
	assert.Equal(t, "Method Not Allowed\n", b)

}
