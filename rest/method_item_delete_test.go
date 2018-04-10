package rest_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/resource/testing/mem"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
)

func TestDeleteItem(t *testing.T) {
	sharedInit := func() *requestTestVars {
		s := mem.NewHandler()
		s.Insert(context.Background(), []*resource.Item{
			{ID: "1", ETag: "a", Payload: map[string]interface{}{"id": "1", "foo": "odd"}},
			{ID: "2", ETag: "b", Payload: map[string]interface{}{"id": "2", "foo": "even"}},
			{ID: "3", ETag: "c", Payload: map[string]interface{}{"id": "3", "foo": "odd"}},
			{ID: "4", ETag: "d", Payload: map[string]interface{}{"id": "4", "foo": "even"}},
			{ID: "5", ETag: "e", Payload: map[string]interface{}{"id": "5", "foo": "odd"}},
		})

		idx := resource.NewIndex()
		idx.Bind("foo", schema.Schema{
			Fields: schema.Fields{
				"id":  {Sortable: true, Filterable: true},
				"foo": {Filterable: true},
			},
		}, s, resource.Conf{AllowedModes: resource.ReadWrite, PaginationDefaultLimit: 2})

		return &requestTestVars{
			Index:   idx,
			Storers: map[string]resource.Storer{"foo": s},
		}
	}
	checkFooIDs := func(ids ...interface{}) requestCheckerFunc {
		return func(t *testing.T, vars *requestTestVars) {
			s := vars.Storers["foo"]
			items, err := s.Find(context.Background(), &query.Query{Sort: query.Sort{{Name: "id", Reversed: false}}})
			if err != nil {
				t.Errorf("s.Find failed: %s", err)
			}
			if el, al := len(ids), len(items.Items); el != al {
				t.Errorf("Expected resource 'foo' to contain %d items, got %d", el, al)
				return
			}
			for i, eid := range ids {
				if aid := items.Items[i].ID; eid != aid {
					el := len(ids)
					t.Errorf("Expected item %d/%d to have ID %q, got ID %q", i+1, el, eid, aid)
				}
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
				return http.NewRequest("DELETE", "/foo/1", nil)
			},
			ResponseCode: http.StatusNotImplemented,
			ResponseBody: `{"code": 501, "message": "No Storage Defined"}`,
		},
		`pathID:not-found`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", `/foo/66`, nil)
			},
			ResponseCode: http.StatusNotFound,
			ResponseBody: `{"code": 404, "message": "Not Found"}`,
			ExtraTest:    checkFooIDs("1", "2", "3", "4", "5"),
		},
		`pathID:found`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", "/foo/2", nil)
			},
			ResponseCode: http.StatusNoContent,
			ResponseBody: ``,
			ExtraTest:    checkFooIDs("1", "3", "4", "5"),
		},
		`pathID:found,filter:invalid-json`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", `/foo/2?filter=invalid`, nil)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code": 422,
				"message": "URL parameters contain error(s)",
				"issues": {
					"filter": ["char 0: expected '{' got 'i'"]
				}}`,
			ExtraTest: checkFooIDs("1", "2", "3", "4", "5"),
		},
		`pathID:found,filter:invalid-json,sort:invalid-field`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", `/foo/2?filter=invalid&sort=invalid`, nil)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code": 422,
				"message": "URL parameters contain error(s)",
				"issues": {
					"filter": ["char 0: expected '{' got 'i'"],
					"sort": ["invalid: unknown sort field"]
				}}`,
			ExtraTest: checkFooIDs("1", "2", "3", "4", "5"),
		},
		`pathID:found,filter:invalid-field,sort:invalid-field`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", `/foo/2?filter={"invalid":true}&sort=invalid`, nil)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code": 422,
				"message": "URL parameters contain error(s)",
				"issues": {
					"filter": ["invalid: unknown query field"],
					"sort": ["invalid: unknown sort field"]
				}}`,
			ExtraTest: checkFooIDs("1", "2", "3", "4", "5"),
		},
		`pathID:found,filter:not-matching`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", `/foo/2?filter={foo:"odd"}`, nil)
			},
			ResponseCode: http.StatusNotFound,
			ResponseBody: `{"code": 404, "message": "Not Found"}`,
			ExtraTest:    checkFooIDs("1", "2", "3", "4", "5"),
		},
		`pathID:found,filter:matching`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", `/foo/2?filter={foo:"even"}`, nil)
			},
			ResponseCode: http.StatusNoContent,
			ResponseBody: ``,
			ExtraTest:    checkFooIDs("1", "3", "4", "5"),
		},
		`pathID:found,header["If-Match"]:not-matching`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				r, err := http.NewRequest("DELETE", "/foo/2", nil)
				if err != nil {
					return nil, err
				}
				r.Header.Set("If-Match", "W/x")
				return r, nil
			},
			ResponseCode: http.StatusPreconditionFailed,
			ResponseBody: `{"code": 412, "message": "Precondition Failed"}`,
			ExtraTest:    checkFooIDs("1", "2", "3", "4", "5"),
		},
		`pathID:found,header["If-Match"]:matching`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				r, err := http.NewRequest("DELETE", "/foo/2", nil)
				if err != nil {
					return nil, err
				}
				r.Header.Set("If-Match", "W/b")
				return r, nil
			},
			ResponseCode: http.StatusNoContent,
			ResponseBody: ``,
			ExtraTest:    checkFooIDs("1", "3", "4", "5"),
		},
	}

	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}
