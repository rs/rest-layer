package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func parseSelector(s string) ([]Field, error) {
	pos := 0
	return parseSelectorExpression([]byte(s), &pos, len(s), false)
}

func TestParseSelectorExpression(t *testing.T) {
	f, err := parseSelector("foo{bar,baz}")
	if assert.NoError(t, err) {
		assert.Equal(t, []Field{{Name: "foo", Fields: []Field{{Name: "bar"}, {Name: "baz"}}}}, f)
	}
	f, err = parseSelector("  foo  \n  { \n bar \t , \n baz \t } \n")
	if assert.NoError(t, err) {
		assert.Equal(t, []Field{{Name: "foo", Fields: []Field{{Name: "bar"}, Field{Name: "baz"}}}}, f)
	}
	f, err = parseSelector("foo{bar{baz}:rab}")
	if assert.NoError(t, err) {
		assert.Equal(t, []Field{{
			Name: "foo",
			Fields: []Field{{
				Name:   "bar",
				Alias:  "rab",
				Fields: []Field{{Name: "baz"}},
			}},
		}}, f)
	}
	f, err = parseSelector("foo(bar=\"baz\")")
	if assert.NoError(t, err) {
		assert.Equal(t, []Field{{Name: "foo", Params: map[string]interface{}{"bar": "baz"}}}, f)
	}
	f, err = parseSelector("foo(bar=\"baz\\\"zab\")")
	if assert.NoError(t, err) {
		assert.Equal(t, []Field{{Name: "foo", Params: map[string]interface{}{"bar": "baz\"zab"}}}, f)
	}
	f, err = parseSelector("foo(bar=-0.2)")
	if assert.NoError(t, err) {
		assert.Equal(t, []Field{{Name: "foo", Params: map[string]interface{}{"bar": -0.2}}}, f)
	}
	f, err = parseSelector("foo(bar = -0.2 , baz = \"zab\")")
	if assert.NoError(t, err) {
		assert.Equal(t, []Field{{Name: "foo", Params: map[string]interface{}{"bar": -0.2, "baz": "zab"}}}, f)
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
