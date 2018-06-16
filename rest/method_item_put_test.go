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
	checkPayload := func(name string, id interface{}, payload map[string]interface{}) requestCheckerFunc {
		return func(t *testing.T, vars *requestTestVars) {
			var item *resource.Item

			s := vars.Storers[name]
			q := query.Query{Predicate: query.Predicate{query.Equal{Field: "id", Value: id}}, Window: &query.Window{Limit: 1}}
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
			ResponseCode: http.StatusOK,
			ResponseBody: `{"id": "2", "foo": "baz"}`,
			ExtraTest:    checkPayload("foo", "2", map[string]interface{}{"id": "2", "foo": "baz"}),
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
