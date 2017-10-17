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
					"bar": {
						Validator: schema.Bool{},
					},
					"foobar": {
						Validator: schema.String{},
					},
				},
			},
		},
	}
	cases := []struct {
		projection string
		parseErr   error
		validErr   error
	}{
		{`parent{child},simple`, nil, nil},
		{`with_params(foo:1)`, nil, nil},
		{`with_params(bar:true)`, nil, nil},
		{`with_params(bar:false)`, nil, nil},
		{`with_params(foobar:"foobar")`, nil, nil},
		{`foo`, nil, errors.New("foo: unknown field")},
		{`simple{child}`, nil, errors.New("simple: field as no children")},
		{`parent{foo}`, nil, errors.New("parent.foo: unknown field")},
		{`simple(foo:1)`, nil, errors.New("simple: params not allowed")},
		{`with_params(baz:1)`, nil, errors.New("with_params: unsupported param name: baz")},
		{`with_params(foo:"a string")`, nil, errors.New("with_params: invalid param `foo' value: not an integer")},
		{`with_params(foo:3.14)`, nil, errors.New("with_params: invalid param `foo' value: not an integer")},
		{`with_params(bar:1)`, nil, errors.New("with_params: invalid param `bar' value: not a Boolean")},
		{`with_params(foobar:true)`, nil, errors.New("with_params: invalid param `foobar' value: not a string")},
		{`*`, nil, nil},
		{`parent,*`, nil, nil},
		{`*,parent`, nil, nil},
		{`*parent`, errors.New("looking for field name at char 2"), nil},
		{`parent*`, errors.New("looking for field name at char 6"), nil},
	}
	for i := range cases {
		tc := cases[i]
		t.Run(tc.projection, func(t *testing.T) {
			pr, err := ParseProjection(tc.projection)
			if !reflect.DeepEqual(err, tc.parseErr) {
				t.Errorf("ParseProjection error = %v, wanted: %v", err, tc.parseErr)
			}
			if err = pr.Validate(s); !reflect.DeepEqual(err, tc.validErr) {
				t.Errorf("Projection.Validate error = %v, wanted: %v", err, tc.validErr)
			}
		})
	}
}
