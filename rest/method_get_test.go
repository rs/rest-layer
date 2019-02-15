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
			ResponseCode:   200,
			ResponseBody:   `[{"foo": "bar"}]`,
			ResponseHeader: http.Header{"Etag": []string{`W/"d41d8cd98f00b204e9800998ecf8427e"`}},
		},
		`fields:foo:minimal`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				r, err := http.NewRequest("GET", `/foo?fields=foo`, nil)
				r.Header.Set("Prefer", "return=minimal")
				return r, err
			},
			ResponseCode:   204,
			ResponseBody:   ``,
			ResponseHeader: http.Header{"Etag": []string{`W/"d41d8cd98f00b204e9800998ecf8427e"`}},
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

func TestGetListFilter(t *testing.T) {
	sharedInit := func() *requestTestVars {
		s := mem.NewHandler()
		s.Insert(context.TODO(), []*resource.Item{
			{ID: "1", Payload: map[string]interface{}{"id": 1, "foo": "bar"}},
			{ID: "2", Payload: map[string]interface{}{"id": 2, "foo": nil}},
			{ID: "3", Payload: map[string]interface{}{"id": 3, "foo2": "bar2"}},
		})

		idx := resource.NewIndex()
		idx.Bind("foo", schema.Schema{
			Fields: schema.Fields{
				"foo": {
					Filterable: true,
					Validator:  &schema.AnyOf{&schema.Null{}, &schema.String{}},
				},
			},
		}, s, resource.DefaultConf)

		return &requestTestVars{
			Index:   idx,
			Storers: map[string]resource.Storer{"foo": s},
		}
	}

	tests := map[string]requestTest{
		`filter:string`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:""}`, nil)
			},
			ResponseCode: 200,
			ResponseBody: `[]`,
		},
		`filter:string2`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:"bar"}`, nil)
			},
			ResponseCode: 200,
			ResponseBody: `[{"foo":"bar","id":1}]`,
		},
		`filter:null`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:null}`, nil)
			},
			ResponseCode: 200,
			ResponseBody: `[{"foo":null,"id":2},{"foo2":"bar2","id":3}]`,
		},
	}
	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}

func TestGetListArray(t *testing.T) {
	sharedInit := func() *requestTestVars {
		s := mem.NewHandler()
		s.Insert(context.TODO(), []*resource.Item{
			{ID: "1", Payload: map[string]interface{}{"id": 1,
				"foo": []interface{}{
					map[string]interface{}{
						"a": "bar",
						"b": 10,
					},
					map[string]interface{}{
						"a": "bar1",
						"b": 101,
					},
				}},
			},
			{ID: "2", Payload: map[string]interface{}{"id": 2,
				"foo": []interface{}{
					map[string]interface{}{
						"a": "bar",
						"b": 20,
					},
				}},
			},
			{ID: "3", Payload: map[string]interface{}{"id": 3,
				"foo": []interface{}{
					map[string]interface{}{
						"a": "baz",
						"b": 30,
						"c": "true",
					},
				}},
			},
		})

		arrayObj := &schema.Object{
			Schema: &schema.Schema{
				Fields: schema.Fields{
					"a": {
						Filterable: true,
						Validator:  &schema.String{},
					},
					"b": {
						Filterable: true,
						Validator:  &schema.Integer{},
					},
					"c": {
						Filterable: true,
						Validator:  &schema.String{},
					},
				},
			},
		}
		idx := resource.NewIndex()
		idx.Bind("foo", schema.Schema{
			Fields: schema.Fields{
				"foo": {
					Filterable: true,
					Validator: &schema.Array{
						Values: schema.Field{
							Validator: arrayObj,
						},
					},
				},
			},
		}, s, resource.DefaultConf)

		return &requestTestVars{
			Index:   idx,
			Storers: map[string]resource.Storer{"foo": s},
		}
	}

	// NOTE: Having an array of objects, only one object needs to match the predicate
	// for the whole record to be returned, including all other objects in that array,
	// that may not match predicate given.
	tests := map[string]requestTest{
		`filter/array:foo.a:not-found`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:{$elemMatch:{a:"mar"}}}`, nil)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `[]`,
		},
		`filter/array:foo.a`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:{$elemMatch:{a:"bar"}}}`, nil)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `[
				{"id":1,"foo":[
					{"a":"bar","b":10},
					{"a":"bar1","b":101}
				]},
				{"id":2,"foo":[{"a":"bar","b":20}]}
			]`,
		},
		`filter/array:foo.a+foo.b`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:{$elemMatch:{a:"bar",b:10}}}`, nil)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `[
				{"id":1,"foo":[
					{"a":"bar","b":10},
					{"a":"bar1","b":101}
				]}
			]`,
		},
		`filter/array:foo.b`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:{$elemMatch:{b:10}}}`, nil)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `[
				{"id":1,"foo":[
					{"a":"bar","b":10},
					{"a":"bar1","b":101}
				]}
			]`,
		},
		`filter/array:foo.a:regex`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:{$elemMatch:{a:{$regex:"az$"}}}}`, nil)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `[
				{"id":3,"foo":[{"a":"baz","b":30,"c":"true"}]}
			]`,
		},
		`filter/array:foo.b:gt`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:{$elemMatch:{b:{$gt:20}}}}`, nil)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `[
				{"id":1,"foo":[
					{"a":"bar","b":10},
					{"a":"bar1","b":101}
				]},
				{"id":3,"foo":[{"a":"baz","b":30,"c":"true"}]}
			]`,
		},
		`filter/array:foo.b:gte`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:{$elemMatch:{b:{$gte:20}}}}`, nil)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `[
				{"id":1,"foo":[
					{"a":"bar","b":10},
					{"a":"bar1","b":101}
				]},
				{"id":2,"foo":[{"a":"bar","b":20}]},
				{"id":3,"foo":[{"a":"baz","b":30,"c":"true"}]}
			]`,
		},
		`filter/array:foo.b:lt`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:{$elemMatch:{b:{$lt:20}}}}`, nil)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `[
				{"id":1,"foo":[
					{"a":"bar","b":10},
					{"a":"bar1","b":101}
				]}
			]`,
		},
		`filter/array:foo.b:lte`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:{$elemMatch:{b:{$lte:20}}}}`, nil)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `[
				{"id":1,"foo":[
					{"a":"bar","b":10},
					{"a":"bar1","b":101}
				]},
				{"id":2,"foo":[{"a":"bar","b":20}]}
			]`,
		},
		`filter/array:foo.b:$in`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:{$elemMatch:{b:{$in:[10,20]}}}}`, nil)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `[
				{"id":1,"foo":[
					{"a":"bar","b":10},
					{"a":"bar1","b":101}
				]},
				{"id":2,"foo":[{"a":"bar","b":20}]}
			]`,
		},
		`filter/array:foo.b:$exists-true`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:{$elemMatch:{c:{$exists:true}}}}`, nil)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `[
				{"id":3,"foo":[{"a":"baz","b":30,"c":"true"}]}
			]`,
		},
		`filter/array:foo.b:$exists-false`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:{$elemMatch:{c:{$exists:false}}}}`, nil)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `[
				{"id":1,"foo":[
					{"a":"bar","b":10},
					{"a":"bar1","b":101}
				]},
				{"id":2,"foo":[{"a":"bar","b":20}]}
			]`,
		},
		`filter/array:foo:not-an-array`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:"mar"}`, nil)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{"code":422,"issues":{"filter":["foo: invalid query expression: not an array"]},"message":"URL parameters contain error(s)"}`,
		},
		`filter/array:foo:invalid-elemMatch`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("GET", `/foo?filter={foo:{$elemMatch:"mar"}}`, nil)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{"code":422,"issues":{"filter":["char 17: foo: $elemMatch: expected '{' got '\"'"]},"message":"URL parameters contain error(s)"}`,
		},
	}
	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}
