package resource

import (
	"errors"
	"fmt"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func parseSelector(s string) ([]Field, error) {
	pos := 0
	return parseSelectorExpression([]byte(s), &pos, len(s), false)
}

func TestParseSelectorExpression(t *testing.T) {
	f, err := parseSelector("foo{bar,baz}")
	if assert.NoError(t, err) {
		assert.Equal(t, []Field{Field{Name: "foo", Fields: []Field{Field{Name: "bar"}, Field{Name: "baz"}}}}, f)
	}
	f, err = parseSelector("  foo  \n  { \n bar \t , \n baz \t } \n")
	if assert.NoError(t, err) {
		assert.Equal(t, []Field{Field{Name: "foo", Fields: []Field{Field{Name: "bar"}, Field{Name: "baz"}}}}, f)
	}
	f, err = parseSelector("foo{bar{baz}:rab}")
	if assert.NoError(t, err) {
		assert.Equal(t, []Field{Field{Name: "foo", Fields: []Field{Field{Name: "bar", Alias: "rab", Fields: []Field{Field{Name: "baz"}}}}}}, f)
	}
	f, err = parseSelector("foo(bar=\"baz\")")
	if assert.NoError(t, err) {
		assert.Equal(t, []Field{Field{Name: "foo", Params: map[string]interface{}{"bar": "baz"}}}, f)
	}
	f, err = parseSelector("foo(bar=\"baz\\\"zab\")")
	if assert.NoError(t, err) {
		assert.Equal(t, []Field{Field{Name: "foo", Params: map[string]interface{}{"bar": "baz\"zab"}}}, f)
	}
	f, err = parseSelector("foo(bar=-0.2)")
	if assert.NoError(t, err) {
		assert.Equal(t, []Field{Field{Name: "foo", Params: map[string]interface{}{"bar": -0.2}}}, f)
	}
	f, err = parseSelector("foo(bar = -0.2 , baz = \"zab\")")
	if assert.NoError(t, err) {
		assert.Equal(t, []Field{Field{Name: "foo", Params: map[string]interface{}{"bar": -0.2, "baz": "zab"}}}, f)
	}
}

func TestParseSelectorExpressionInvalid(t *testing.T) {
	_, err := parseSelector("foo{bar,baz")
	assert.EqualError(t, err, "looking for `}' at char 11")
	_, err = parseSelector("foo:bar{baz}")
	assert.EqualError(t, err, "looking for `,` and got `{' at char 7")
	_, err = parseSelector("foo:{bar}")
	assert.EqualError(t, err, "looking for field alias at char 4")
	_, err = parseSelector("foo{}")
	assert.EqualError(t, err, "looking for field name at char 4")
	_, err = parseSelector("{foo}")
	assert.EqualError(t, err, "looking for field name at char 0")
	_, err = parseSelector(",foo")
	assert.EqualError(t, err, "looking for field name at char 0")
	_, err = parseSelector("f oo")
	assert.EqualError(t, err, "invalid char at 2")
	_, err = parseSelector("foo}")
	assert.EqualError(t, err, "looking for field name and got `}' at char 3")
	_, err = parseSelector("foo()")
	assert.EqualError(t, err, "looking for parameter name at char 4")
	_, err = parseSelector("foo(bar baz)")
	assert.EqualError(t, err, "looking for = at char 8")
	_, err = parseSelector("foo(bar")
	assert.EqualError(t, err, "looking for = at char 7")
	_, err = parseSelector("foo(bar=\"baz)")
	assert.EqualError(t, err, "looking for \" at char 13")
	_, err = parseSelector("foo(bar=0a)")
	assert.EqualError(t, err, "looking for `,' or ')' at char 9")
	_, err = parseSelector("foo(bar=@toto)")
	assert.EqualError(t, err, "looking for value at char 8")
}

func TestValidateSelector(t *testing.T) {
	s := schema.Schema{
		"parent": schema.Field{
			Schema: &schema.Schema{
				"child": schema.Field{},
			},
		},
		"simple": schema.Field{},
		"with_params": schema.Field{
			Params: &schema.Params{
				Validators: map[string]schema.FieldValidator{
					"foo": schema.Integer{},
				},
			},
		},
	}

	assert.NoError(t, validateSelector([]Field{{Name: "parent", Fields: []Field{{Name: "child"}}}}, s))
	assert.NoError(t, validateSelector([]Field{{Name: "with_params", Params: map[string]interface{}{"foo": 1}}}, s))

	assert.EqualError(t,
		validateSelector([]Field{{Name: "foo"}}, s),
		"foo: unknown field")
	assert.EqualError(t,
		validateSelector([]Field{{Name: "simple", Fields: []Field{{Name: "child"}}}}, s),
		"simple: field as no children")
	assert.EqualError(t,
		validateSelector([]Field{{Name: "parent", Fields: []Field{{Name: "foo"}}}}, s),
		"parent.foo: unknown field")
	assert.EqualError(t,
		validateSelector([]Field{{Name: "simple", Params: map[string]interface{}{"foo": 1}}}, s),
		"simple: params not allowed")
	assert.EqualError(t,
		validateSelector([]Field{{Name: "with_params", Params: map[string]interface{}{"bar": 1}}}, s),
		"with_params: unsupported param name: bar")
	assert.EqualError(t,
		validateSelector([]Field{{Name: "with_params", Params: map[string]interface{}{"foo": "a string"}}}, s),
		"with_params: invalid param `foo' value: not an integer")
}

func TestApplySelector(t *testing.T) {
	s := schema.Schema{
		"parent": schema.Field{
			Schema: &schema.Schema{
				"child": schema.Field{},
			},
		},
		"simple": schema.Field{},
		"with_params": schema.Field{
			Params: &schema.Params{
				Handler: func(value interface{}, params map[string]interface{}) (interface{}, error) {
					if val, found := params["foo"]; found {
						if val == -1 {
							return nil, errors.New("some error")
						}
						return fmt.Sprintf("param is %d", val), nil
					}
					return "no param", nil
				},
				Validators: map[string]schema.FieldValidator{
					"foo": schema.Integer{},
				},
			},
		},
	}

	// Basic filtering
	p, err := applySelector([]Field{{Name: "parent", Fields: []Field{{Name: "child"}}}}, s,
		map[string]interface{}{"parent": map[string]interface{}{"child": "value"}, "simple": "value"}, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]interface{}{"parent": map[string]interface{}{"child": "value"}}, p)
	}
	// Alias on both parent and child
	p, err = applySelector([]Field{{Name: "parent", Alias: "p", Fields: []Field{{Name: "child", Alias: "c"}}}}, s,
		map[string]interface{}{"parent": map[string]interface{}{"child": "value"}}, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]interface{}{"p": map[string]interface{}{"c": "value"}}, p)
	}
	// Param call with valid value
	p, err = applySelector([]Field{{Name: "with_params", Params: map[string]interface{}{"foo": 1}}}, s,
		map[string]interface{}{"with_params": "value"}, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]interface{}{"with_params": "param is 1"}, p)
	}
	// If no param, handler is not called at all
	p, err = applySelector([]Field{{Name: "with_params"}}, s,
		map[string]interface{}{"with_params": "value"}, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]interface{}{"with_params": "value"}, p)
	}
	// Param call with valid value rejected by the handler
	p, err = applySelector([]Field{{Name: "with_params", Params: map[string]interface{}{"foo": -1}}}, s,
		map[string]interface{}{"with_params": "value"}, nil)
	assert.EqualError(t, err, "with_params: some error")
	assert.Nil(t, p)
	// Param call on a field with no param set
	p, err = applySelector([]Field{{Name: "simple", Params: map[string]interface{}{"foo": "bar"}}}, s,
		map[string]interface{}{"simple": "value"}, nil)
	assert.EqualError(t, err, "simple: params not allowed")
	assert.Nil(t, p)
	// Deep field lookup on a field with no child
	p, err = applySelector([]Field{{Name: "simple", Fields: []Field{{Name: "child"}}}}, s,
		map[string]interface{}{"simple": "value"}, nil)
	assert.EqualError(t, err, "simple: field as no children")
	assert.Nil(t, p)
	// Deep field lookup on a field with invalid payload (no dict)
	p, err = applySelector([]Field{{Name: "parent", Fields: []Field{{Name: "child"}}}}, s,
		map[string]interface{}{"parent": "value"}, nil)
	assert.EqualError(t, err, "parent: invalid value: not a dict")
	assert.Nil(t, p)
	// Deep field lookup with invalid child
	p, err = applySelector([]Field{{Name: "parent", Fields: []Field{{Name: "child", Params: map[string]interface{}{"foo": "bar"}}}}}, s,
		map[string]interface{}{"parent": map[string]interface{}{"child": "value"}}, nil)
	assert.EqualError(t, err, "parent.child: params not allowed")
	assert.Nil(t, p)
}
