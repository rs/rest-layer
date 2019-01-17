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
			"array": {
				Validator: &schema.Array{
					Values: schema.Field{
						Validator: &schema.Object{
							Schema: &schema.Schema{
								Fields: schema.Fields{"child": {}},
							},
						},
					},
				},
			},
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
			"connection": {
				Validator: &schema.Connection{
					Path:  "cnx",
					Field: "ref",
					Validator: schema.Schema{Fields: schema.Fields{
						"id":   {},
						"name": {},
					}},
				},
			},
		},
	}
	cases := []struct {
		projection string
		err        error
	}{
		{`array{child}`, nil},
		{`array{*}`, nil},
		{`array{foo}`, errors.New("array.foo: unknown field")},
		{`parent{child},simple`, nil},
		{`with_params(foo:1)`, nil},
		{`with_params(bar:true)`, nil},
		{`with_params(bar:false)`, nil},
		{`with_params(foobar:"foobar")`, nil},
		{`foo`, errors.New("foo: unknown field")},
		{`simple{child}`, errors.New("simple: field has no children")},
		{`parent{foo}`, errors.New("parent.foo: unknown field")},
		{`simple(foo:1)`, errors.New("simple: params not allowed")},
		{`with_params(baz:1)`, errors.New("with_params: unsupported param name: baz")},
		{`with_params(foo:"a string")`, errors.New("with_params: invalid param `foo' value: not an integer")},
		{`with_params(foo:3.14)`, errors.New("with_params: invalid param `foo' value: not an integer")},
		{`with_params(bar:1)`, errors.New("with_params: invalid param `bar' value: not a Boolean")},
		{`with_params(foobar:true)`, errors.New("with_params: invalid param `foobar' value: not a string")},
		{`*`, nil},
		{`*,*`, nil},
		{`z:*`, errors.New("*: can't have an alias")},
		{`simple{*}`, errors.New("simple: field has no children")},
		{`parent{*}`, nil},
		{`parent{foo}`, errors.New("parent.foo: unknown field")},
		{`*,parent{*}`, nil},
		{`*,parent{z:*}`, errors.New("parent.*: can't have an alias")},
		{`*,parent{child{*}}`, errors.New("parent.child: field has no children")},
		{`connection{name}`, nil},
		{`connection{*}`, nil},
		{`connection{foo}`, errors.New("connection.foo: unknown field")},
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
