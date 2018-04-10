package rest_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/resource/testing/mem"
	"github.com/rs/rest-layer/schema"
)

func TestGetItem(t *testing.T) {
	sharedInit := func() *requestTestVars {
		s := mem.NewHandler()
		s.Insert(context.TODO(), []*resource.Item{
			{ID: "1", Payload: map[string]interface{}{"id": "1", "foo": "bar"}},
			{ID: "2", Payload: map[string]interface{}{"id": "2", "foo": "baz"}},
			{ID: "3", Payload: map[string]interface{}{"id": "3"}},
		})

		idx := resource.NewIndex()
		idx.Bind("foo", schema.Schema{
			Fields: schema.Fields{
				"id":  {Filterable: true},
				"foo": {Filterable: true},
			},
		}, s, resource.DefaultConf)

		return &requestTestVars{
			Index:   idx,
			Storers: map[string]resource.Storer{"foo": s},
		}
	}

	tests := map[string]requestTest{
		"fields:invalid": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo/1?fields=invalid", nil)
			},
			ResponseCode: 422,
			ResponseBody: `{
				"code": 422,
				"message": "URL parameters contain error(s)",
				"issues": {
					"fields": ["invalid: unknown field"]
				}
			}`,
		},
		// sort is currently allowed by rest-layer on a single item fetch by ID.
		"sort:invalid": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo/1?sort=invalid", nil)
			},
			ResponseCode: 422,
			ResponseBody: `{
				"code": 422,
				"message": "URL parameters contain error(s)",
				"issues": {
					"sort": ["invalid: unknown sort field"]
				}
			}`,
		},
		// filter is currently allowed by rest-layer on a single item fetch by ID.
		"filter:invalid": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo/1?filter=invalid", nil)
			},
			ResponseCode: 422,
			ResponseBody: `{
				"code": 422,
				"message": "URL parameters contain error(s)",
				"issues": {
					"filter": ["char 0: expected '{' got 'i'"]
				}
			}`,
		},
		"pathID:found": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo/2", nil)

			},
			ResponseCode: 200,
			ResponseBody: `{"id": "2", "foo": "baz"}`,
		},
		`pathID:found,filter:match`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo/2?filter={foo:"baz"}`, nil)
			},
			ResponseCode: 200,
			ResponseBody: `{"id": "2", "foo": "baz"}`,
		},
		`pathID:found,filter:no-match`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo/2?filter={foo:"bar"}`, nil)
			},
			ResponseCode: 404,
			ResponseBody: `{"code": 404, "message": "Not Found"}`,
		},
	}

	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}

func TestGetItemConditionally(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	sharedInit := func() *requestTestVars {
		s := mem.NewHandler()
		s.Insert(context.TODO(), []*resource.Item{
			{ID: "1", ETag: "a", Updated: now, Payload: map[string]interface{}{"id": "1", "foo": "bar"}},
			{ID: "2", ETag: "b", Updated: yesterday, Payload: map[string]interface{}{"id": "2", "foo": "baz"}},
		})

		idx := resource.NewIndex()
		idx.Bind("foo", schema.Schema{}, s, resource.DefaultConf)

		return &requestTestVars{
			Index:   idx,
			Storers: map[string]resource.Storer{"foo": s},
		}
	}

	tests := map[string]requestTest{
		`header["If-None-Match"]:matching`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				r, err := http.NewRequest("GET", `/foo/2`, nil)
				if err != nil {
					return nil, err
				}
				r.Header.Set("If-None-Match", "W/b")
				return r, nil
			},
			ResponseCode: http.StatusNotModified,
			ResponseBody: ``,
		},
		`header["If-None-Match"]:not-matching`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				r, err := http.NewRequest("GET", `/foo/2`, nil)
				if err != nil {
					return nil, err
				}
				r.Header.Set("If-None-Match", "W/x")
				return r, nil
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"id": "2", "foo": "baz"}`,
		},
		`header["If-Modified-Since"]:invalid`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				r, err := http.NewRequest("GET", `/foo/1`, nil)
				if err != nil {
					return nil, err
				}
				r.Header.Set("If-Modified-Since", "invalid")
				return r, nil
			},
			ResponseCode: http.StatusBadRequest,
			ResponseBody: `{"code": 400, "message": "Invalid If-Modified-Since header"}`,
		},
		`header["If-Modified-Since"]:not-matching`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				r, err := http.NewRequest("GET", `/foo/2`, nil)
				if err != nil {
					return nil, err
				}
				r.Header.Set("If-Modified-Since", now.Format(time.RFC1123))
				return r, nil
			},
			ResponseCode: http.StatusNotModified,
			ResponseBody: ``,
		},
		`header["If-Modified-Since"]:exact-match`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				r, err := http.NewRequest("GET", `/foo/2`, nil)
				if err != nil {
					return nil, err
				}
				r.Header.Set("If-Modified-Since", yesterday.Format(time.RFC1123))
				return r, nil
			},
			ResponseCode: http.StatusNotModified,
			ResponseBody: ``,
		},
		`header["If-Modified-Since"]:matching`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				r, err := http.NewRequest("GET", `/foo/1`, nil)
				if err != nil {
					return nil, err
				}
				r.Header.Set("If-Modified-Since", yesterday.Format(time.RFC1123))
				return r, nil
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"id": "1", "foo": "bar"}`,
		},
	}

	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}
func TestGetItemFieldHandler(t *testing.T) {
	sharedInit := func() *requestTestVars {
		s := mem.NewHandler()
		s.Insert(context.TODO(), []*resource.Item{
			{ID: "1", Payload: map[string]interface{}{"id": "1", "foo": "bar"}},
		})

		idx := resource.NewIndex()
		idx.Bind("foo", schema.Schema{
			Fields: schema.Fields{
				"foo": {
					Params: map[string]schema.Param{
						"bar": {},
					},
					Handler: func(ctx context.Context, value interface{}, params map[string]interface{}) (interface{}, error) {
						if s, _ := params["bar"].(string); s != "baz" {
							return nil, errors.New("error")
						}
						return "baz", nil
					},
				},
			},
		}, s, resource.DefaultConf)

		return &requestTestVars{
			Index:   idx,
			Storers: map[string]resource.Storer{"foo": s},
		}
	}

	tests := map[string]requestTest{
		`fields:foo(bar:invalid)`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo/1?fields=foo(bar:"invalid")`, nil)
			},
			ResponseCode: 520,
			ResponseBody: `{"code": 520, "message": "foo: error"}`,
		},
		`fields:foo(bar:baz)`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo/1?fields=foo(bar:"baz")`, nil)
			},
			ResponseCode: 200,
			ResponseBody: `{"foo": "baz"}`,
		},
		`fields:foo`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo/1?fields=foo`, nil)
			},
			ResponseCode: 200,
			ResponseBody: `{"foo": "bar"}`,
		},
	}
	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}

func TestHandlerGetItemNoStorage(t *testing.T) {
	sharedInit := func() *requestTestVars {
		idx := resource.NewIndex()
		idx.Bind("foo", schema.Schema{}, nil, resource.DefaultConf)
		return &requestTestVars{
			Index: idx,
		}
	}

	tc := requestTest{
		Init: sharedInit,
		NewRequest: func() (*http.Request, error) {
			return http.NewRequest("GET", "/foo/1", nil)
		},
		ResponseCode: http.StatusNotImplemented,
		ResponseBody: `{"code": 501, "message": "No Storage Defined"}`,
	}
	tc.Test(t)
}
