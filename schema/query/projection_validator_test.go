package query

import (
	"errors"
	"reflect"
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestProjectionValidate(t *testing.T) {
	s := schema.Schema{
		Fields: schema.Fields{
			"parent": {
				Schema: &schema.Schema{
					Fields: schema.Fields{"child": {}},
				},
			},
			"simple": schema.Field{},
			"with_params": {
				Params: schema.Params{
					"foo": {
						Validator: schema.Integer{},
					},
				},
			},
		},
	}
	cases := []struct {
		projection string
		err        error
	}{
		{`parent{child},simple`, nil},
		{`with_params(foo:1)`, nil},
		{`foo`, errors.New("foo: unknown field")},
		{`simple{child}`, errors.New("simple: field as no children")},
		{`parent{foo}`, errors.New("parent.foo: unknown field")},
		{`simple(foo:1)`, errors.New("simple: params not allowed")},
		{`with_params(bar:1)`, errors.New("with_params: unsupported param name: bar")},
		{`with_params(foo:"a string")`, errors.New("with_params: invalid param `foo' value: not an integer")},
	}
	for i := range cases {
		tc := cases[i]
		t.Run(tc.projection, func(t *testing.T) {
			pr, err := ParseProjection(tc.projection)
			if err != nil {
				t.Errorf("ParseProjection unexpected error: %v", err)
			}
			if err = pr.Validate(s); !reflect.DeepEqual(err, tc.err) {
				t.Errorf("Projection.Validate error = %v, wanted: %v", err, tc.err)
			}
		})
	}
}
