package query

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/rs/rest-layer/schema"
)

type resource struct {
	validator schema.Validator
}

func (r resource) Find(ctx context.Context, query *Query) ([]map[string]interface{}, error) {
	return nil, errors.New("not implemented")
}
func (r resource) MultiGet(ctx context.Context, ids []interface{}) ([]map[string]interface{}, error) {
	return nil, errors.New("not implemented")
}
func (r resource) SubResource(ctx context.Context, path string) (Resource, error) {
	return nil, errors.New("not implemented")
}
func (r resource) Validator() schema.Validator {
	return r.validator
}

func TestProjectionEval(t *testing.T) {
	s := schema.Schema{
		Fields: schema.Fields{
			"parent": {
				Schema: &schema.Schema{
					Fields: schema.Fields{
						"child": {},
					},
				},
			},
			"simple": schema.Field{},
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
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			pr, err := ParseProjection(tc.projection)
			if err != nil {
				t.Errorf("ParseProjection unexpected error: %v", err)
			}
			if err = pr.Validate(s); err != nil {
				t.Errorf("Projection.Validate unexpected error: %v", err)
			}
			var payload map[string]interface{}
			err = json.Unmarshal([]byte(tc.payload), &payload)
			if err != nil {
				t.Errorf("Invalid JSON payload: %v", err)
			}
			payload, err = pr.Eval(ctx, payload, resource{s})
			if !reflect.DeepEqual(err, tc.err) {
				t.Errorf("Eval return error: %v, wanted: %v", err, tc.err)
			}
			if err != nil {
				return
			}
			got, _ := json.Marshal(payload)
			if string(got) != tc.want {
				t.Errorf("Eval = %v, wanted: %v", string(got), tc.want)
			}
		})
	}
}
