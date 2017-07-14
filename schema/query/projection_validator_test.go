package query

import (
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
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

	assert.NoError(t, Projection{{Name: "parent", Children: Projection{{Name: "child"}}}}.Validate(s))
	assert.NoError(t, Projection{{Name: "with_params", Params: map[string]interface{}{"foo": 1}}}.Validate(s))

	assert.EqualError(t,
		Projection{{Name: "foo"}}.Validate(s),
		"foo: unknown field")
	assert.EqualError(t,
		Projection{{Name: "simple", Children: Projection{{Name: "child"}}}}.Validate(s),
		"simple: field as no children")
	assert.EqualError(t,
		Projection{{Name: "parent", Children: Projection{{Name: "foo"}}}}.Validate(s),
		"parent.foo: unknown field")
	assert.EqualError(t,
		Projection{{Name: "simple", Params: map[string]interface{}{"foo": 1}}}.Validate(s),
		"simple: params not allowed")
	assert.EqualError(t,
		Projection{{Name: "with_params", Params: map[string]interface{}{"bar": 1}}}.Validate(s),
		"with_params: unsupported param name: bar")
	assert.EqualError(t,
		Projection{{Name: "with_params", Params: map[string]interface{}{"foo": "a string"}}}.Validate(s),
		"with_params: invalid param `foo' value: not an integer")
}

// func TestLookupSetSelector(t *testing.T) {
// 	l := NewLookup()
// 	v := schema.Schema{
// 		Fields: schema.Fields{
// 			"foo": {
// 				Schema: &schema.Schema{
// 					Fields: schema.Fields{
// 						"bar": {},
// 					},
// 				},
// 			},
// 			"baz": {},
// 		},
// 	}
// 	err := l.SetSelector(`foo{bar},baz`, v)
// 	assert.NoError(t, err)
// 	err = l.SetSelector(`foo,`, v)
// 	assert.EqualError(t, err, "looking for field name at char 4")
// 	err = l.SetSelector(`bar`, v)
// 	assert.EqualError(t, err, "bar: unknown field")
// }
