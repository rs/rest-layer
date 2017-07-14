package rest_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
	"github.com/stretchr/testify/assert"
)

func TestHandlerPostList(t *testing.T) {
	tests := map[string]requestTest{
		"OK": {
			Init: func() *requestTestVars {
				i := resource.NewIndex()
				s := mem.NewHandler()
				i.Bind("foo", schema.Schema{Fields: schema.Fields{
					"id":  {OnInit: func(ctx context.Context, v interface{}) interface{} { return "1" }},
					"foo": {},
					"bar": {},
				}}, s, resource.DefaultConf)
				return &requestTestVars{Index: i, Storers: map[string]resource.Storer{"foo": s}}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/foo", bytes.NewBufferString(`{"foo": "bar"}`))
			},
			ResponseCode: 201,
			ResponseBody: `{"foo":"bar","id":"1"}`,
			ExtraTest: func(t *testing.T, vars *requestTestVars) {
				q := &query.Query{
					Predicate: query.Predicate{query.Equal{Field: "id", Value: "1"}},
					Window:    &query.Window{Limit: 1},
				}
				s, ok := vars.Storers["foo"]
				if !assert.True(t, ok) {
					return
				}
				l, err := s.Find(context.TODO(), q)
				assert.NoError(t, err)
				if assert.Len(t, l.Items, 1) {
					assert.Equal(t, map[string]interface{}{"id": "1", "foo": "bar"}, l.Items[0].Payload)
				}
			},
		},
		"BadPayload": {
			Init: func() *requestTestVars {
				index := resource.NewIndex()
				index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/test", bytes.NewBufferString(`{invalid json`))
			},
			ResponseCode: http.StatusBadRequest,
			ResponseBody: `{
				"code": 400,
				"message": "Malformed body: invalid character 'i' looking for beginning of object key string"
			}`,
		},
		"InvalIDQueryFields": {
			Init: func() *requestTestVars {
				index := resource.NewIndex()
				index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/test?fields=invalid", bytes.NewBufferString(`{}`))
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code": 422,
				"message": "Invalid ` + "`fields`" + ` parameter: invalid: unknown field"
			}`,
		},
		"Dup": {
			Init: func() *requestTestVars {
				index := resource.NewIndex()
				s := mem.NewHandler()
				s.Insert(context.TODO(), []*resource.Item{
					{ID: "1", Payload: map[string]interface{}{"id": "1", "foo": "bar"}},
					{ID: "2", Payload: map[string]interface{}{"id": "2", "foo": "bar"}},
					{ID: "3", Payload: map[string]interface{}{"id": "3", "foo": "bar"}},
				})
				index.Bind("test", schema.Schema{Fields: schema.Fields{
					"id":  {},
					"foo": {},
				}}, s, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/test", bytes.NewBufferString(`{"id": "2", "foo": "baz"}`))
			},
			ResponseCode: http.StatusConflict,
			ResponseBody: `{"code":409,"message":"Conflict"}`,
		},
		"New": {
			Init: func() *requestTestVars {
				index := resource.NewIndex()
				s := mem.NewHandler()
				index.Bind("test", schema.Schema{Fields: schema.Fields{
					"id":  {},
					"foo": {},
				}}, s, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/test", bytes.NewBufferString(`{"id": "2", "foo": "baz"}`))
			},
			ResponseCode:   http.StatusCreated,
			ResponseBody:   `{"id":"2","foo":"baz"}`,
			ResponseHeader: http.Header{"Content-Location": []string{"/test/2"}},
		},
		"InvalidField": {
			Init: func() *requestTestVars {
				index := resource.NewIndex()
				s := mem.NewHandler()
				index.Bind("test", schema.Schema{Fields: schema.Fields{
					"id": {},
				}}, s, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/test", bytes.NewBufferString(`{"id": "2", "foo": "baz"}`))
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code": 422,
				"message": "Document contains error(s)",
				"issues": {"foo": ["invalid field"]}
			}`,
		},
		"MissingID": {
			Init: func() *requestTestVars {
				index := resource.NewIndex()
				s := mem.NewHandler()
				index.Bind("test", schema.Schema{Fields: schema.Fields{
					"id": {},
				}}, s, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/test", bytes.NewBufferString(`{}`))
			},
			// FIXME: The HTTP 520 code is usually used for protocol errors, and
			// seems unaprporiate. This should most likely be a 422 error.
			ResponseCode: 520,
			ResponseBody: `{
				"code": 520,
				"message": "Missing ID field"
			}`,
		},
		"NoStorage": {
			// FIXME: For NoStorage, it's probably better to error early (during Bind).
			Init: func() *requestTestVars {
				index := resource.NewIndex()
				index.Bind("test", schema.Schema{Fields: schema.Fields{
					"id": {},
				}}, nil, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/test", bytes.NewBufferString(`{"id":1}`))
			},
			ResponseCode: http.StatusNotImplemented,
			ResponseBody: `{
				"code": 501,
				"message": "No Storage Defined"
			}`,
		},
		"WithReferenceNotFound": {
			Init: func() *requestTestVars {
				s := mem.NewHandler()
				index := resource.NewIndex()
				index.Bind("foo", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
				index.Bind("bar", schema.Schema{Fields: schema.Fields{
					"id":  {},
					"foo": {Validator: &schema.Reference{Path: "foo"}},
				}}, s, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/bar", bytes.NewBufferString(`{"id": "1", "foo": "nonexisting"}`))
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code": 422,
				"message": "Document contains error(s)",
				"issues": {"foo": ["Not Found"]}
			}`,
		},
		"WithReferenceNoStorage": {
			// FIXME: For NoStorage, it's probably better to error early (during Bind).
			Init: func() *requestTestVars {
				s := mem.NewHandler()
				index := resource.NewIndex()
				index.Bind("foo", schema.Schema{Fields: schema.Fields{"id": {}}}, nil, resource.DefaultConf)
				index.Bind("bar", schema.Schema{Fields: schema.Fields{
					"id":  {},
					"foo": {Validator: &schema.Reference{Path: "foo"}},
				}}, s, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/bar", bytes.NewBufferString(`{"id": "1", "foo": "1"}`))
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code": 422,
				"message": "Document contains error(s)",
				"issues": {"foo": ["No Storage Defined"]}
			}`,
		},
		"WithReference": {
			Init: func() *requestTestVars {
				s := mem.NewHandler()
				s.Insert(context.Background(), []*resource.Item{{ID: "ref", Payload: map[string]interface{}{"id": "ref"}}})
				index := resource.NewIndex()
				index.Bind("foo", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
				index.Bind("bar", schema.Schema{Fields: schema.Fields{
					"id":  {},
					"foo": {Validator: &schema.Reference{Path: "foo"}},
				}}, s, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/bar", bytes.NewBufferString(`{"id": "1", "foo": "ref"}`))
			},
			ResponseCode: http.StatusCreated,
			ResponseBody: `{"id": "1", "foo": "ref"}`,
		},
		"WithSubSchemaReference": {
			Init: func() *requestTestVars {
				s := mem.NewHandler()
				s.Insert(context.Background(), []*resource.Item{{ID: "ref", Payload: map[string]interface{}{"id": "ref"}}})
				index := resource.NewIndex()
				index.Bind("foo", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
				index.Bind("bar", schema.Schema{Fields: schema.Fields{
					"id": {},
					"sub": {Schema: &schema.Schema{Fields: schema.Fields{
						"foo": {Validator: &schema.Reference{Path: "foo"}},
					}}},
				}}, s, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/bar", bytes.NewBufferString(`{"id": "1", "sub": {"foo": "ref"}}`))
			},
			ResponseCode: http.StatusCreated,
			ResponseBody: `{"id": "1", "sub": {"foo": "ref"}}`,
		},
		"WithSubSchemaObjectReference": {
			Init: func() *requestTestVars {
				s := mem.NewHandler()
				s.Insert(context.Background(), []*resource.Item{{ID: "ref", Payload: map[string]interface{}{"id": "ref"}}})
				index := resource.NewIndex()
				index.Bind("foo", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
				index.Bind("bar", schema.Schema{Fields: schema.Fields{
					"id": {},
					"sub": {Validator: &schema.Object{Schema: &schema.Schema{Fields: schema.Fields{
						"foo": {Validator: &schema.Reference{Path: "foo"}},
					}}}},
				}}, s, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/bar", bytes.NewBufferString(`{"id": "1", "sub": {"foo": "ref"}}`))
			},
			ResponseCode: http.StatusCreated,
			ResponseBody: `{"id": "1", "sub": {"foo": "ref"}}`,
		},
		"WithArraySchemaReferenceNotFound": {
			Init: func() *requestTestVars {
				s := mem.NewHandler()
				s.Insert(context.Background(), []*resource.Item{
					{ID: "ref1", Payload: map[string]interface{}{"id": "ref1"}},
				})
				index := resource.NewIndex()
				index.Bind("foo", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
				index.Bind("bar", schema.Schema{Fields: schema.Fields{
					"id": {},
					"foos": {Validator: &schema.Array{
						ValuesValidator: &schema.Reference{Path: "foo"},
					}},
				}}, s, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/bar", bytes.NewBufferString(`{"id": "1", "foos": ["ref1", "ref2"]}`))
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code": 422,
				"message": "Document contains error(s)",
				"issues": {"foos":["invalid value at #2: Not Found"]}
			}`,
		},
		"WithArraySchemaReference": {
			Init: func() *requestTestVars {
				s := mem.NewHandler()
				s.Insert(context.Background(), []*resource.Item{
					{ID: "ref1", Payload: map[string]interface{}{"id": "ref1"}},
					{ID: "ref2", Payload: map[string]interface{}{"id": "ref2"}},
				})
				index := resource.NewIndex()
				index.Bind("foo", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
				index.Bind("bar", schema.Schema{Fields: schema.Fields{
					"id": {},
					"foos": {Validator: &schema.Array{
						ValuesValidator: &schema.Reference{Path: "foo"},
					}},
				}}, s, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/bar", bytes.NewBufferString(`{"id": "1", "foos": ["ref1", "ref2"]}`))
			},
			ResponseCode: http.StatusCreated,
			ResponseBody: `{"id": "1", "foos": ["ref1", "ref2"]}`,
		},
	}
	for name, tt := range tests {
		tt := tt // capture range variable.
		t.Run(name, tt.Test)
	}
}
func TestHandlerPostListWithInvalidReference(t *testing.T) {
	s := mem.NewHandler()
	index := resource.NewIndex()
	index.Bind("bar", schema.Schema{Fields: schema.Fields{
		"id":  {},
		"foo": {Validator: &schema.Reference{Path: "invalid"}},
	}}, s, resource.DefaultConf)

	h, err := rest.NewHandler(index)
	assert.Error(t, err, "bar: schema compilation error: foo: can't find resource 'invalid'", "rest.NewHandler(index)")
	assert.Nil(t, h, "rest.NewHandler(index)")
}
