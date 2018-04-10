package rest_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/resource/testing/mem"
	"github.com/rs/rest-layer/schema"
)

func TestGetListInvalidQuery(t *testing.T) {
	sharedInit := func() *requestTestVars {
		s := mem.NewHandler()

		idx := resource.NewIndex()
		idx.Bind("foo", schema.Schema{}, s, resource.Conf{AllowedModes: resource.ReadWrite})

		return &requestTestVars{
			Index:   idx,
			Storers: map[string]resource.Storer{"foo": s},
		}
	}

	tests := map[string]requestTest{
		"fields:invalid": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo?fields=invalid", nil)
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
		"sort:invalid": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo?sort=invalid", nil)
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
		"filter:invalid": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo?filter=invalid", nil)
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
		"page:invalid": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo?page=invalid", nil)
			},
			ResponseCode: 422,
			ResponseBody: `{
				"code": 422,
				"message": "URL parameters contain error(s)",
				"issues": {
					"page": ["must be positive integer"]
				}
			}`,
		},
		"page:-1": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo?page=-1", nil)
			},
			ResponseCode: 422,
			ResponseBody: `{
				"code": 422,
				"message": "URL parameters contain error(s)",
				"issues": {
					"page": ["must be positive integer"]
				}
			}`,
		},
		"page:2,limit:missing": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo?page=2", nil)
			},
			ResponseCode: 422,
			ResponseBody: `{
				"code": 422,
				"message": "URL parameters contain error(s)",
				"issues": {
					"limit": ["required when page is set and there is no resource default"]
				}
			}`,
		},
		"limit:invalid": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo?limit=invalid", nil)
			},
			ResponseCode: 422,
			ResponseBody: `{
				"code": 422,
				"message": "URL parameters contain error(s)",
				"issues": {
					"limit": ["must be positive integer"]
				}
			}`,
		},
		"limit:-1": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo?limit=-1", nil)
			},
			ResponseCode: 422,
			ResponseBody: `{
				"code": 422,
				"message": "URL parameters contain error(s)",
				"issues": {
					"limit": ["must be positive integer"]
				}
			}`,
		},
	}

	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}

func TestGetListPagination(t *testing.T) {
	sharedInit := func() *requestTestVars {
		s := mem.NewHandler()
		s.Insert(context.TODO(), []*resource.Item{
			{ID: "1", Payload: map[string]interface{}{"id": "1"}},
			{ID: "2", Payload: map[string]interface{}{"id": "2"}},
			{ID: "3", Payload: map[string]interface{}{"id": "3"}},
			{ID: "4", Payload: map[string]interface{}{"id": "4"}},
			{ID: "5", Payload: map[string]interface{}{"id": "5"}},
		})

		idx := resource.NewIndex()
		idx.Bind("foo", schema.Schema{}, s, resource.DefaultConf)

		return &requestTestVars{
			Index:   idx,
			Storers: map[string]resource.Storer{"foo": s},
		}
	}

	tests := map[string]requestTest{
		"page:2,limit:2": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo?page=2&limit=2", nil)
			},
			ResponseCode: 200,
			ResponseBody: `[{"id": "3"}, {"id": "4"}]`,
			ResponseHeader: http.Header{
				"X-Offset": []string{"2"},
				"X-Total":  []string{"5"},
			},
		},
		"page:3,limit:2": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo?page=3&limit=2", nil)
			},
			ResponseCode: 200,
			ResponseBody: `[{"id": "5"}]`,
			ResponseHeader: http.Header{
				"X-Offset": []string{"4"},
				"X-Total":  []string{"5"},
			},
		},
		"skip:1,page:2,limit:2": {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", "/foo?skip=1&page=2&limit=2", nil)
			},
			ResponseCode: 200,
			ResponseBody: `[{"id": "4"},{"id": "5"}]`,
			ResponseHeader: http.Header{
				"X-Offset": []string{"3"},
				"X-Total":  []string{"5"},
			},
		},
	}
	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}
func TestGetListFieldHandler(t *testing.T) {
	sharedInit := func() *requestTestVars {
		s := mem.NewHandler()
		s.Insert(context.TODO(), []*resource.Item{
			{ID: "1", Payload: map[string]interface{}{"id": 1, "foo": "bar"}},
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
		`fields:foo`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?fields=foo`, nil)
			},
			ResponseCode: 200,
			ResponseBody: `[{"foo": "bar"}]`,
		},
		`fields:foo(bar:baz)`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?fields=foo(bar:"baz")`, nil)
			},
			ResponseCode: 200,
			ResponseBody: `[{"foo": "baz"}]`,
		},
		`fields:foo(bar:invalid)`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?fields=foo(bar:"invalid")`, nil)
			},
			// FIXME: 520 is mostly used for HTTP protocol errors, and seams inappropriate.
			// should probably use 422, or possibly 500.
			ResponseCode: 520,
			ResponseBody: `{
				"code": 520,
				"message": "foo: error"
			}`,
		},
	}
	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}
