package query

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/rs/rest-layer/internal/testutil"
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
	cnxShema := schema.Schema{Fields: schema.Fields{
		"name": {},
		"ref":  {},
	}}

	r := resource{
		validator: schema.Schema{Fields: schema.Fields{
			"id":     {},
			"simple": {},
			"parent": {
				Schema: &schema.Schema{
					Fields: schema.Fields{
						"child":  {},
						"child2": {},
					},
				},
			},
			"dict": {
				Validator: &schema.Dict{
					Values: schema.Field{
						Validator: &schema.Reference{
							Path: "cnx",
						},
					},
				},
			},
			"arrayObject": {
				Validator: &schema.Array{
					Values: schema.Field{
						Validator: &schema.Object{
							Schema: &schema.Schema{
								Fields: schema.Fields{
									"child":  {},
									"child2": {},
								},
							},
						},
					},
				},
			},
			"reference": {
				Validator: &schema.Reference{
					Path:            "cnx",
					SchemaValidator: cnxShema,
				},
			},
			"arrayReference": {
				Validator: &schema.Array{
					Values: schema.Field{
						Validator: &schema.Reference{
							Path:            "cnx",
							SchemaValidator: cnxShema,
						},
					},
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
					"bar": {Validator: schema.Bool{}},
				},
				Handler: func(ctx context.Context, value interface{}, params map[string]interface{}) (interface{}, error) {
					if val, found := params["foo"]; found {
						if val == -1 {
							return nil, errors.New("some error")
						}
						return fmt.Sprintf("param foo is %d", val), nil
					}
					if val, found := params["bar"]; found {
						return fmt.Sprintf("param bar is %t", val), nil
					}
					return "no param", nil
				},
			},
		}},
		subResources: map[string]resource{
			"cnx": resource{
				validator: cnxShema,
				payloads: map[string]map[string]interface{}{
					"1": map[string]interface{}{"id": "1", "name": "first"},
					"2": map[string]interface{}{"id": "2", "name": "second", "ref": "a"},
					"3": map[string]interface{}{"id": "3", "name": "third", "ref": "b"},
					"4": map[string]interface{}{"id": "4", "name": "forth", "ref": "a"},
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
			`{"with_params":"param foo is 1"}`,
		},
		{
			"Parmeters",
			`with_params(bar:true)`,
			`{"with_params":"value"}`,
			nil,
			`{"with_params":"param bar is true"}`,
		},
		{
			"Parmeters",
			`with_params(bar:false)`,
			`{"with_params":"value"}`,
			nil,
			`{"with_params":"param bar is false"}`,
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
			`reference`,
			`{"reference":"100"}`,
			nil,
			`{"reference":"100"}`,
		},
		{
			"Reference/Field",
			`reference{name}`,
			`{"reference":"2"}`,
			nil,
			`{"reference":{"name":"second"}}`,
		},
		{
			"Reference/Field-non-existant",
			`reference{name}`,
			`{"reference":"100"}`,
			nil,
			`{"reference":null}`,
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
		{
			"Star",
			`*`,
			`{"parent":{"child":"value"},"simple":"value","reference":"2"}`,
			nil,
			`{"parent":{"child":"value"},"simple":"value","reference":"2"}`,
		},
		{
			"Star/Expand-one-level",
			`*{*}`,
			`{"parent":{"child":"value"},"reference":"2"}`,
			nil,
			`{"parent":{"child":"value"},"reference":{"id":"2","name":"second","ref":"a"}}`,
		},
		{
			"Star/Expand-invalid",
			`*{*}`,
			`{"parent":{"child":"value"},"simple":"value","reference":"2"}`,
			errors.New("simple: field has no children"),
			``,
		},
		{
			"Star/Double",
			`*,*`,
			`{"parent":{"child":"value"},"simple":"value","reference":"2"}`,
			errors.New("only one * in projection allowed"),
			``,
		},
		{
			"Star/Rename",
			`*,s:simple`,
			`{"parent":{"child":"value"},"simple":"value","reference":"2"}`,
			nil,
			`{"parent":{"child":"value"},"reference":"2","s":"value","simple":"value"}`,
		},
		{
			"Star/Parent",
			`parent{*}`,
			`{"parent":{"child":"value"},"simple":"value","reference":"2"}`,
			nil,
			`{"parent":{"child":"value"}}`,
		},
		{
			"Reference/Expand",
			`reference{*}`,
			`{"parent":{"child":"value"},"simple":"value","reference":"2"}`,
			nil,
			`{"reference":{"id":"2","name":"second","ref":"a"}}`,
		},
		{
			"Reference/Expand-double",
			`reference{*,*}`,
			`{"parent":{"child":"value"},"simple":"value","reference":"2"}`,
			errors.New("reference: error applying Projection on sub-field: only one * in projection allowed"),
			``,
		},
		{
			"Reference/Expand-mixed",
			`reference{*},*`,
			`{"parent":{"child":"value"},"simple":"value","reference":"2"}`,
			nil,
			`{"parent":{"child":"value"},"reference":{"id":"2","name":"second","ref":"a"},"simple":"value"}`,
		},
		{
			"Reference/Expand-rename",
			`reference{*,n:name}`,
			`{"parent":{"child":"value"},"simple":"value","reference":"2"}`,
			nil,
			`{"reference":{"id":"2","n":"second","name":"second","ref":"a"}}`,
		},
		{
			"ArrayOfObject",
			`arrayObject`,
			`{"arrayObject":[{"child":"foo", "child2":"bar"},{"child":"foo2"}]}`,
			nil,
			`{"arrayObject":[{"child":"foo","child2":"bar"},{"child":"foo2"}]}`,
		},
		{
			"ArrayOfObject/Expand-field",
			`arrayObject{child}`,
			`{"arrayObject":[{"child":"foo", "child2":"bar"},{"child":"foo2"}]}`,
			nil,
			`{"arrayObject":[{"child":"foo"},{"child":"foo2"}]}`,
		},
		{
			"ArrayOfObject/Expand",
			`arrayObject{*}`,
			`{"arrayObject":[{"child":"foo","child2":"bar"},{"child":"foo2"}]}`,
			nil,
			`{"arrayObject":[{"child":"foo","child2":"bar"},{"child":"foo2"}]}`,
		},
		{
			"ArrayOfObject/Expand-rename",
			`arrayObject{c:child}`,
			`{"arrayObject":[{"child":"foo","child2":"bar"},{"child":"foo2"}]}`,
			nil,
			`{"arrayObject":[{"c":"foo"},{"c":"foo2"}]}`,
		},
		{
			"ArrayOfObject/Expand-and-rename",
			`arrayObject{*,c:child}`,
			`{"arrayObject":[{"child":"foo","child2":"bar"},{"child":"foo2"}]}`,
			nil,
			`{"arrayObject":[{"c":"foo","child":"foo","child2":"bar"},{"c":"foo2","child":"foo2"}]}`,
		},
		{
			"ArrayReference/Field-non-existant",
			`arrayReference{name}`,
			`{"arrayReference":["100"]}`,
			nil,
			`{"arrayReference":[]}`,
		},
		{
			"ArrayReference/Empty",
			`arrayReference{name}`,
			`{"arrayReference":[]}`,
			nil,
			`{"arrayReference":[]}`,
		},
		{
			"ArrayReference/Expand",
			`arrayReference{*}`,
			`{"arrayReference":["2","3"]}`,
			nil,
			`{"arrayReference":[{"id":"2","name":"second","ref":"a"},{"id":"3","name":"third","ref":"b"}]}`,
		},
		{
			"ArrayReference/Expand-non-existant",
			`arrayReference{*}`,
			`{"arrayReference":["100","101"]}`,
			nil,
			`{"arrayReference":[]}`,
		},
		{
			"ArrayReference/Field",
			`arrayReference{name}`,
			`{"arrayReference":["2"]}`,
			nil,
			`{"arrayReference":[{"name":"second"}]}`,
		},
		{
			"ArrayReference/Field-many",
			`arrayReference{name}`,
			`{"arrayReference":["2","3"]}`,
			nil,
			`{"arrayReference":[{"name":"second"},{"name":"third"}]}`,
		},
		{
			"ArrayReference/Field-many-non-existant",
			`arrayReference{name}`,
			`{"arrayReference":["2","100","3"]}`,
			nil,
			`{"arrayReference":[{"name":"second"},{"name":"third"}]}`,
		},
		{
			"Dict",
			`dict`,
			`{"dict":{"x":"2"}}`,
			nil,
			`{"dict":{"x":"2"}}`,
		},
		{
			"Dict/Multiple",
			`dict{x,y}`,
			`{"dict":{"x":"2","y":"3"}}`,
			nil,
			`{"dict":{"x":"2","y":"3"}}`,
		},
		{
			"Dict/Field-unknown",
			`dict{z}`,
			`{"dict":{"x":"2"}}`,
			nil,
			`{"dict":{}}`,
		},
		{
			"Dict/Field-unknown-mixed",
			`dict{x,z}`,
			`{"dict":{"x":"2"}}`,
			nil,
			`{"dict":{"x":"2"}}`,
		},
		{
			"Dict/Expand",
			`dict{*}`,
			`{"dict":{"x":"2","y":"3"}}`,
			nil,
			`{"dict":{"x":"2","y":"3"}}`,
		},
		{
			"Dict/Expand-all",
			`dict{*{*}}`,
			`{"dict":{"x":"2","y":"3"}}`,
			nil,
			`{"dict":{"x":{"id":"2","name":"second","ref":"a"},"y":{"id":"3","name":"third","ref":"b"}}}`,
		},
		{
			"Dict/Expand-reference",
			`dict{*,y{*}}`,
			`{"dict":{"x":"2","y":"3"}}`,
			nil,
			`{"dict":{"x":"2","y":{"id":"3","name":"third","ref":"b"}}}`,
		},
		{
			"Dict/Expand-reference-partial",
			`dict{*,y{name}}`,
			`{"dict":{"x":"2","y":"3"}}`,
			nil,
			`{"dict":{"x":"2","y":{"name":"third"}}}`,
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
			testutil.JSONEq(t, []byte(tc.want), []byte(got))
		})
	}
}
