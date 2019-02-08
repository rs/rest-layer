package rest_test

import (
	"bytes"
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/resource/testing/mem"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
)

func checkPayload(name string, id interface{}, payload map[string]interface{}) requestCheckerFunc {
	return func(t *testing.T, vars *requestTestVars) {
		var item *resource.Item

		s := vars.Storers[name]
		q := query.Query{Predicate: query.Predicate{&query.Equal{Field: "id", Value: id}}, Window: &query.Window{Limit: 1}}
		if items, err := s.Find(context.Background(), &q); err != nil {
			t.Errorf("s.Find failed: %s", err)
			return
		} else if len(items.Items) != 1 {
			t.Errorf("item with ID %v not found", id)
			return
		} else {
			item = items.Items[0]
		}
		if !reflect.DeepEqual(payload, item.Payload) {
			t.Errorf("Unexpected stored payload for item %v:\nexpect: %#v\ngot: %#v", id, payload, item.Payload)
		}
	}
}

func TestPutItem(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	sharedInit := func() *requestTestVars {
		s1 := mem.NewHandler()
		s1.Insert(context.Background(), []*resource.Item{
			{ID: "1", ETag: "a", Updated: now, Payload: map[string]interface{}{"id": "1", "foo": "odd", "bar": "baz"}},
			{ID: "2", ETag: "b", Updated: yesterday, Payload: map[string]interface{}{"id": "2", "foo": "even", "bar": "baz"}},
			{ID: "3", ETag: "c", Updated: yesterday, Payload: map[string]interface{}{"id": "3", "foo": "odd", "bar": "baz"}},
		})
		s2 := mem.NewHandler()
		s2.Insert(context.Background(), []*resource.Item{
			{ID: "1", ETag: "d", Updated: now, Payload: map[string]interface{}{"id": "1", "foo": "3"}},
		})

		idx := resource.NewIndex()
		foo := idx.Bind("foo", schema.Schema{
			Fields: schema.Fields{
				"id":  {Sortable: true, Filterable: true},
				"foo": {Filterable: true},
				"bar": {Filterable: true},
			},
		}, s1, resource.DefaultConf)
		foo.Bind("sub", "foo", schema.Schema{
			Fields: schema.Fields{
				"id":  {Sortable: true, Filterable: true},
				"foo": {Filterable: true, Validator: &schema.Reference{Path: "foo"}},
			},
		}, s2, resource.DefaultConf)

		return &requestTestVars{
			Index:   idx,
			Storers: map[string]resource.Storer{"foo": s1, "foo.sub": s2},
		}
	}

	tests := map[string]requestTest{
		`NoStorage`: {
			// FIXME: For NoStorage, it's probably better to error early (during Bind).
			Init: func() *requestTestVars {
				index := resource.NewIndex()
				index.Bind("foo", schema.Schema{}, nil, resource.DefaultConf)
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"id": "3"}`))
				return http.NewRequest("PUT", "/foo/1", body)
			},
			ResponseCode: http.StatusNotImplemented,
			ResponseBody: `{"code": 501, "message": "No Storage Defined"}`,
		},
		`CreateModeNotAllowed`: {
			Init: func() *requestTestVars {
				s := mem.NewHandler()
				index := resource.NewIndex()
				index.Bind("foo", schema.Schema{}, s, resource.Conf{AllowedModes: []resource.Mode{resource.Replace}})
				return &requestTestVars{Index: index}
			},
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "bar"}`))
				return http.NewRequest("PUT", "/foo/66", body)
			},
			ResponseCode: http.StatusMethodNotAllowed,
			ResponseBody: `{"code": 405, "message": "Method Not Allowed"}`,
		},
		`ReplaceModeNotAllowed`: {
			Init: func() *requestTestVars {
				s := mem.NewHandler()
				s.Insert(context.Background(), []*resource.Item{
					{ID: "1", Payload: map[string]interface{}{"id": "1", "foo": "bar"}},
				})
				index := resource.NewIndex()
				index.Bind("foo", schema.Schema{}, s, resource.Conf{AllowedModes: []resource.Mode{resource.Create}})
				return &requestTestVars{Index: index, Storers: map[string]resource.Storer{"foo": s}}
			},
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				return http.NewRequest("PUT", "/foo/1", body)
			},
			ResponseCode: http.StatusMethodNotAllowed,
			ResponseBody: `{"code": 405, "message": "Method Not Allowed"}`,
			ExtraTest:    checkPayload("foo", "1", map[string]interface{}{"id": "1", "foo": "bar"}),
		},
		`pathID:not-found,body:valid`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				return http.NewRequest("PUT", `/foo/66`, body)
			},
			ResponseCode: http.StatusCreated,
			ResponseBody: `{"id": "66", "foo": "baz"}`,
			ExtraTest:    checkPayload("foo", "66", map[string]interface{}{"id": "66", "foo": "baz"}),
		},
		`pathID:found,body:invalid-json`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`invalid`))
				return http.NewRequest("PUT", "/foo/2", body)
			},
			ResponseCode: http.StatusBadRequest,
			ResponseBody: `{
				"code": 400,
				"message": "Malformed body: invalid character 'i' looking for beginning of value"
			}`,
			ExtraTest: checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "even", "bar": "baz"}),
		},
		`pathID:found,body:invalid-field`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"invalid": true}`))
				return http.NewRequest("PUT", "/foo/2", body)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code":422,
				"message": "Document contains error(s)",
				"issues": {
					"invalid": ["invalid field"]
				}
			}`,
			ExtraTest: checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "even", "bar": "baz"}),
		},
		`pathID:found,body:alter-id`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"id": "3"}`))
				return http.NewRequest("PUT", "/foo/2", body)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code":422,
				"message": "Cannot change document ID"
			}`,
			ExtraTest: checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "even", "bar": "baz"}),
		},
		`pathID:found,body:valid`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				return http.NewRequest("PUT", "/foo/2", body)
			},
			ResponseCode:   http.StatusOK,
			ResponseBody:   `{"id": "2", "foo": "baz"}`,
			ResponseHeader: http.Header{"Etag": []string{`W/"b89c2acfea8a49933a3387f0e3fb0527"`}},
			ExtraTest:      checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "baz"}),
		},
		`pathID:found,body:valid:minimal`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				r, err := http.NewRequest("PUT", "/foo/2", body)
				r.Header.Set("Prefer", "return=minimal")
				return r, err
			},
			ResponseCode:   http.StatusNoContent,
			ResponseBody:   ``,
			ResponseHeader: http.Header{"Etag": []string{`W/"b89c2acfea8a49933a3387f0e3fb0527"`}},
			ExtraTest:      checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "baz"}),
		},
		`pathID:found,body:valid,fields:invalid`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				return http.NewRequest("PUT", "/foo/2?fields=invalid", body)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code": 422,
				"message": "URL parameters contain error(s)",
				"issues": {
					"fields": ["invalid: unknown field"]
				}
			}`,
			ExtraTest: checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "even", "bar": "baz"}),
		},
		`pathID:found,body:valid,fields:valid`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				return http.NewRequest("PUT", "/foo/2?fields=foo", body)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"foo": "baz"}`,
			ExtraTest:    checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "baz"}),
		},
		`pathID:found,body:valid,header["If-Match"]:not-matching`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				r, err := http.NewRequest("PUT", "/foo/2", body)
				if err != nil {
					return nil, err
				}
				r.Header.Set("If-Match", "W/x")
				return r, nil
			},
			ResponseCode: http.StatusPreconditionFailed,
			ResponseBody: `{"code": 412, "message": "Precondition Failed"}`,
			ExtraTest:    checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "even", "bar": "baz"}),
		},
		`pathID:found,body:valid,header["If-Match"]:matching`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				r, err := http.NewRequest("PUT", "/foo/2", body)
				if err != nil {
					return nil, err
				}
				r.Header.Set("If-Match", "W/b")
				return r, nil
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"id": "2", "foo": "baz"}`,
			ExtraTest:    checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "baz"}),
		},
		`pathID:found,body:valid,header["If-Unmodified-Since"]:invalid`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				r, err := http.NewRequest("PUT", "/foo/1", body)
				if err != nil {
					return nil, err
				}
				r.Header.Set("If-Unmodified-Since", "invalid")
				return r, nil
			},
			ResponseCode: http.StatusBadRequest,
			ResponseBody: `{"code": 400, "message": "Invalid If-Unmodified-Since header"}`,
			ExtraTest:    checkPayload("foo", "1", map[string]interface{}{"id": "1", "foo": "odd", "bar": "baz"}),
		},
		`pathID:found,body:valid,header["If-Unmodified-Since"]:not-matching`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				r, err := http.NewRequest("PUT", "/foo/1", body)
				if err != nil {
					return nil, err
				}
				r.Header.Set("If-Unmodified-Since", yesterday.Format(time.RFC1123))
				return r, nil
			},
			ResponseCode: http.StatusPreconditionFailed,
			ResponseBody: `{"code": 412, "message": "Precondition Failed"}`,
			ExtraTest:    checkPayload("foo", "1", map[string]interface{}{"id": "1", "foo": "odd", "bar": "baz"}),
		},
		`parentPathID:found,pathID:found,body:alter-parent-id`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "2"}`))
				r, err := http.NewRequest("PUT", "/foo/3/sub/1", body)
				if err != nil {
					return nil, err
				}
				return r, nil
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"id": "1", "foo": "2"}`,
			ExtraTest:    checkPayload("foo.sub", "1", map[string]interface{}{"id": "1", "foo": "2"}),
		},
		`parentPathID:found,pathID:found,body:no-parent-id`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{}`))
				r, err := http.NewRequest("PUT", "/foo/3/sub/1", body)
				if err != nil {
					return nil, err
				}
				return r, nil
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"id": "1", "foo": "3"}`,
			ExtraTest:    checkPayload("foo.sub", "1", map[string]interface{}{"id": "1", "foo": "3"}),
		},
	}

	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}

func TestPutItemDefault(t *testing.T) {
	now := time.Now()

	sharedInit := func() *requestTestVars {
		s1 := mem.NewHandler()
		s1.Insert(context.Background(), []*resource.Item{
			{ID: "1", ETag: "a", Updated: now, Payload: map[string]interface{}{"id": "1", "foo": "odd"}},
			{ID: "2", ETag: "b", Updated: now, Payload: map[string]interface{}{"id": "2", "foo": "odd", "bar": "value"}},
		})
		idx := resource.NewIndex()
		idx.Bind("foo", schema.Schema{
			Fields: schema.Fields{
				"id":  {Sortable: true, Filterable: true},
				"foo": {Filterable: true},
				"bar": {Filterable: true, Default: "default"},
			},
		}, s1, resource.DefaultConf)
		return &requestTestVars{
			Index:   idx,
			Storers: map[string]resource.Storer{"foo": s1},
		}
	}

	tests := map[string]requestTest{
		`pathID:not-found,body:valid,default:missing`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				return http.NewRequest("PUT", "/foo/66", body)
			},
			ResponseCode: http.StatusCreated,
			ResponseBody: `{"id": "66", "foo": "baz", "bar": "default"}`,
			ExtraTest:    checkPayload("foo", "66", map[string]interface{}{"id": "66", "foo": "baz", "bar": "default"}),
		},
		`pathID:not-found,body:valid,default:set`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz", "bar": "value"}`))
				return http.NewRequest("PUT", "/foo/66", body)
			},
			ResponseCode: http.StatusCreated,
			ResponseBody: `{"id": "66", "foo": "baz", "bar": "value"}`,
			ExtraTest:    checkPayload("foo", "66", map[string]interface{}{"id": "66", "foo": "baz", "bar": "value"}),
		},
		`pathID:found,body:valid,default:missing`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				return http.NewRequest("PUT", "/foo/1", body)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"id": "1", "foo": "baz"}`,
			ExtraTest:    checkPayload("foo", "1", map[string]interface{}{"id": "1", "foo": "baz"}),
		},
		`pathID:found,body:valid,default:set`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz", "bar": "value"}`))
				return http.NewRequest("PUT", "/foo/1", body)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"id": "1", "foo": "baz", "bar": "value"}`,
			ExtraTest:    checkPayload("foo", "1", map[string]interface{}{"id": "1", "foo": "baz", "bar": "value"}),
		},
		`pathID:found,body:valid,default:delete`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				return http.NewRequest("PUT", "/foo/2", body)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"id": "2", "foo": "baz"}`,
			ExtraTest:    checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "baz"}),
		},
	}

	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}

func TestPutItemRequired(t *testing.T) {
	now := time.Now()

	sharedInit := func() *requestTestVars {
		s1 := mem.NewHandler()
		s1.Insert(context.Background(), []*resource.Item{
			{ID: "1", ETag: "a", Updated: now, Payload: map[string]interface{}{"id": "1", "foo": "odd"}},
			{ID: "2", ETag: "b", Updated: now, Payload: map[string]interface{}{"id": "2", "foo": "odd", "bar": "original"}},
		})
		idx := resource.NewIndex()
		idx.Bind("foo", schema.Schema{
			Fields: schema.Fields{
				"id":  {Sortable: true, Filterable: true},
				"foo": {Filterable: true},
				"bar": {Filterable: true, Required: true},
			},
		}, s1, resource.DefaultConf)
		return &requestTestVars{
			Index:   idx,
			Storers: map[string]resource.Storer{"foo": s1},
		}
	}

	tests := map[string]requestTest{
		`pathID:not-found,body:valid,required:missing`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				return http.NewRequest("PUT", "/foo/66", body)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code": 422,
				"message": "Document contains error(s)",
				"issues": {
					"bar": ["required"]
				}
			}`,
		},
		`pathID:not-found,body:valid,required:set`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz", "bar": "value"}`))
				return http.NewRequest("PUT", "/foo/1", body)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"id": "1", "foo": "baz", "bar": "value"}`,
			ExtraTest:    checkPayload("foo", "1", map[string]interface{}{"id": "1", "foo": "baz", "bar": "value"}),
		},
		`pathID:found,body:valid,required:missing`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				return http.NewRequest("PUT", "/foo/1", body)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code": 422,
				"message": "Document contains error(s)",
				"issues": {
					"bar": ["required"]
				}
			}`,
		},
		`pathID:found,body:valid,required:change`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz", "bar": "value1"}`))
				return http.NewRequest("PUT", "/foo/2", body)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"id": "2", "foo": "baz", "bar": "value1"}`,
			ExtraTest:    checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "baz", "bar": "value1"}),
		},
		`pathID:found,body:valid,required:delete`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				return http.NewRequest("PUT", "/foo/2", body)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code": 422,
				"message": "Document contains error(s)",
				"issues": {
					"bar": ["required"]
				}
			}`,
		},
	}

	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}

func TestPutItemRequiredDefault(t *testing.T) {
	now := time.Now()

	sharedInit := func() *requestTestVars {
		s1 := mem.NewHandler()
		s1.Insert(context.Background(), []*resource.Item{
			{ID: "1", ETag: "a", Updated: now, Payload: map[string]interface{}{"id": "1", "foo": "odd"}},
			{ID: "2", ETag: "b", Updated: now, Payload: map[string]interface{}{"id": "2", "foo": "odd", "bar": "original"}},
		})
		idx := resource.NewIndex()
		idx.Bind("foo", schema.Schema{
			Fields: schema.Fields{
				"id":  {Sortable: true, Filterable: true},
				"foo": {Filterable: true},
				"bar": {Filterable: true, Required: true, Default: "default"},
			},
		}, s1, resource.DefaultConf)
		return &requestTestVars{
			Index:   idx,
			Storers: map[string]resource.Storer{"foo": s1},
		}
	}

	tests := map[string]requestTest{
		`pathID:not-found,body:valid,required-default:missing`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				return http.NewRequest("PUT", "/foo/66", body)
			},
			ResponseCode: http.StatusCreated,
			ResponseBody: `{"id": "66", "foo": "baz", "bar": "default"}`,
			ExtraTest:    checkPayload("foo", "66", map[string]interface{}{"id": "66", "foo": "baz", "bar": "default"}),
		},
		`pathID:not-found,body:valid,required-default:set`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz", "bar": "value"}`))
				return http.NewRequest("PUT", "/foo/1", body)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"id": "1", "foo": "baz", "bar": "value"}`,
			ExtraTest:    checkPayload("foo", "1", map[string]interface{}{"id": "1", "foo": "baz", "bar": "value"}),
		},
		`pathID:found,body:valid,required-default:missing`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				return http.NewRequest("PUT", "/foo/1", body)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code": 422,
				"message": "Document contains error(s)",
				"issues": {
					"bar": ["required"]
				}
			}`,
		},
		`pathID:found,body:valid,required-default:change`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz", "bar": "value"}`))
				return http.NewRequest("PUT", "/foo/2", body)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"id": "2", "foo": "baz", "bar": "value"}`,
			ExtraTest:    checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "baz", "bar": "value"}),
		},
		`pathID:found,body:valid,required-default:delete`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz"}`))
				return http.NewRequest("PUT", "/foo/2", body)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"id": "2", "foo": "baz", "bar": "default"}`,
			ExtraTest:    checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "baz", "bar": "default"}),
		},
	}

	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}

func TestPutItemReadOnly(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	timeOldStr := "2018-01-03T00:00:00+02:00"
	timeOld, _ := time.Parse(time.RFC3339, timeOldStr)

	sharedInit := func() *requestTestVars {
		s1 := mem.NewHandler()
		s1.Insert(context.Background(), []*resource.Item{
			{ID: "1", ETag: "a", Updated: yesterday, Payload: map[string]interface{}{"id": "1", "foo": "odd"}},
			{ID: "2", ETag: "b", Updated: yesterday, Payload: map[string]interface{}{"id": "2", "foo": "odd", "zar": "old"}},
			// Storer will persist `schema.Time{}` as `time.Time` type.
			{ID: "3", ETag: "c", Updated: yesterday, Payload: map[string]interface{}{"id": "3", "foo": "odd", "tar": timeOld}},
		})

		idx := resource.NewIndex()
		idx.Bind("foo", schema.Schema{
			Fields: schema.Fields{
				"id":  {Sortable: true, Filterable: true},
				"foo": {Filterable: true},
				"zar": {ReadOnly: true},
				"tar": {ReadOnly: true, Validator: &schema.Time{}},
			},
		}, s1, resource.DefaultConf)

		return &requestTestVars{
			Index:   idx,
			Storers: map[string]resource.Storer{"foo": s1},
		}
	}

	tests := map[string]requestTest{
		`put:read-only:string:new`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz", "zar":"old"}`))
				return http.NewRequest("PUT", `/foo/1`, body)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{"code":422,"issues":{"zar":["read-only"]},"message":"Document contains error(s)"}`,
			ExtraTest:    checkPayload("foo", "1", map[string]interface{}{"id": "1", "foo": "odd"}),
		},
		`put:read-only:string:old`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "baz", "zar":"old"}`))
				return http.NewRequest("PUT", `/foo/2`, body)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"foo":"baz","id":"2","zar":"old"}`,
			ExtraTest:    checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "baz", "zar": "old"}),
		},
		`put:read-only:time:old`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "odd", "tar":"2018-01-03T00:00:00+02:00"}`))
				return http.NewRequest("PUT", `/foo/3`, body)
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{"foo":"odd","id":"3","tar":"2018-01-03T00:00:00+02:00"}`,
			ExtraTest:    checkPayload("foo", "3", map[string]interface{}{"id": "3", "foo": "odd", "tar": timeOld}),
		},
		`put:read-only:time:new`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				body := bytes.NewReader([]byte(`{"foo": "odd", "tar":"2018-01-03T11:11:11+02:00"}`))
				return http.NewRequest("PUT", `/foo/3`, body)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{"code":422,"issues":{"tar":["read-only"]},"message":"Document contains error(s)"}`,
			ExtraTest:    checkPayload("foo", "3", map[string]interface{}{"id": "3", "foo": "odd", "tar": timeOld}),
		},
	}

	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}
