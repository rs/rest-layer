package query

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/rs/rest-layer/schema"
)

type resource struct {
	validator    schema.Validator
	subResources map[string]resource
	payloads     map[string]map[string]interface{}
}

func (r resource) Find(ctx context.Context, query *Query) ([]map[string]interface{}, error) {
	payloads := []map[string]interface{}{}
	ids := make([]string, 0, len(r.payloads))
	for id := range r.payloads {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		p := r.payloads[id]
		if query.Predicate.Match(p) {
			payloads = append(payloads, p)
		}
	}
	return payloads, nil
}
func (r resource) MultiGet(ctx context.Context, ids []interface{}) ([]map[string]interface{}, error) {
	payloads := make([]map[string]interface{}, len(ids))
	for i, id := range ids {
		if p, found := r.payloads[id.(string)]; found {
			payloads[i] = p
		}
	}
	return payloads, nil
}
func (r resource) SubResource(ctx context.Context, path string) (Resource, error) {
	if sr, found := r.subResources[path]; found {
		return sr, nil
	}
	return nil, errors.New("resource not found")
}
func (r resource) Validator() schema.Validator {
	return r.validator
}

func TestProjectionEval(t *testing.T) {
	r := resource{
		validator: schema.Schema{Fields: schema.Fields{
			"id":     {},
			"simple": {},
			"parent": {
				Schema: &schema.Schema{
					Fields: schema.Fields{
						"child": {},
					},
				},
			},
			"reference": {
				Validator: &schema.Reference{
					Path: "cnx",
				},
			},
			"connection": {
				Validator: &schema.Connection{
					Path:  "cnx",
					Field: "ref",
				},
			},
			"with_params": {
				Params: schema.Params{
					"foo": {Validator: schema.Integer{}},
				},
				Handler: func(ctx context.Context, value interface{}, params map[string]interface{}) (interface{}, error) {
					if val, found := params["foo"]; found {
						if val == -1 {
							return nil, errors.New("some error")
						}
						return fmt.Sprintf("param is %d", val), nil
					}
					return "no param", nil
				},
			},
		}},
		subResources: map[string]resource{
			"cnx": resource{
				validator: schema.Schema{Fields: schema.Fields{
					"name": {},
					"ref":  {},
				}},
				payloads: map[string]map[string]interface{}{
					"1": map[string]interface{}{"name": "first"},
					"2": map[string]interface{}{"name": "second", "ref": "a"},
					"3": map[string]interface{}{"name": "third", "ref": "b"},
					"4": map[string]interface{}{"name": "forth", "ref": "a"},
				},
			},
		},
	}
	cases := []struct {
		name       string
		projection string
		payload    string
		err        error
		want       string
	}{
		{
			"All",
			``,
			`{"parent":{"child":"value"},"simple":"value"}`,
			nil,
			`{"parent":{"child":"value"},"simple":"value"}`,
		},
		{
			"Basic",
			`parent{child}`,
			`{"parent":{"child":"value"},"simple":"value"}`,
			nil,
			`{"parent":{"child":"value"}}`,
		},
		{
			"Aliasing",
			`p:parent{c:child}`,
			`{"parent":{"child":"value"}}`,
			nil,
			`{"p":{"c":"value"}}`,
		},
		{
			"Parmeters",
			`with_params(foo:1)`,
			`{"with_params":"value"}`,
			nil,
			`{"with_params":"param is 1"}`,
		},
		{
			"Parmeters/NoParam", // Handler is not called.
			`with_params`,
			`{"with_params":"value"}`,
			nil,
			`{"with_params":"value"}`,
		},
		{
			"Parmeters/InvalidParam",
			`with_params(foo:-1)`,
			`{"with_params":"value"}`,
			errors.New("with_params: some error"),
			``,
		},
		{
			"InvalidPayload",
			`parent{child}`,
			`{"parent":"value"}`,
			errors.New("parent: invalid value: not a dict"),
			``,
		},
		{
			"Reference",
			`reference{name}`,
			`{"reference":"2"}`,
			nil,
			`{"reference":{"name":"second"}}`,
		},
		{
			"Connection#1",
			`connection{name}`,
			`{"id":"a","simple":"foo"}`,
			nil,
			`{"connection":[{"name":"second"},{"name":"forth"}]}`,
		},
		{
			"Connection#2",
			`connection{name}`,
			`{"id":"b","simple":"foo"}`,
			nil,
			`{"connection":[{"name":"third"}]}`,
		},
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			pr, err := ParseProjection(tc.projection)
			if err != nil {
				t.Errorf("ParseProjection unexpected error: %v", err)
			}
			if err = pr.Validate(r.validator); err != nil {
				t.Errorf("Projection.Validate unexpected error: %v", err)
			}
			var payload map[string]interface{}
			err = json.Unmarshal([]byte(tc.payload), &payload)
			if err != nil {
				t.Errorf("Invalid JSON payload: %v", err)
			}
			payload, err = pr.Eval(ctx, payload, r)
			if !reflect.DeepEqual(err, tc.err) {
				t.Errorf("Eval return error: %v, wanted: %v", err, tc.err)
			}
			if err != nil {
				return
			}
			got, _ := json.Marshal(payload)
			if string(got) != tc.want {
				t.Errorf("Eval:\ngot:  %v\nwant: %v", string(got), tc.want)
			}
		})
	}
}
