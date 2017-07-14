package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProjectionExpression(t *testing.T) {
	f, err := ParseProjection("foo{bar,baz}")
	if assert.NoError(t, err) {
		assert.Equal(t, Projection{{Name: "foo", Children: Projection{{Name: "bar"}, {Name: "baz"}}}}, f)
	}
	f, err = ParseProjection("  foo  \n  { \n bar \t , \n baz \t } \n")
	if assert.NoError(t, err) {
		assert.Equal(t, Projection{{Name: "foo", Children: Projection{{Name: "bar"}, {Name: "baz"}}}}, f)
	}
	f, err = ParseProjection("rab:foo{bar{baz}}")
	if assert.NoError(t, err) {
		assert.Equal(t, Projection{{
			Name:  "foo",
			Alias: "rab",
			Children: Projection{{
				Name:     "bar",
				Children: Projection{{Name: "baz"}},
			}},
		}}, f)
	}
	f, err = ParseProjection("foo{rab:bar{baz}}")
	if assert.NoError(t, err) {
		assert.Equal(t, Projection{{
			Name: "foo",
			Children: Projection{{
				Name:     "bar",
				Alias:    "rab",
				Children: Projection{{Name: "baz"}},
			}},
		}}, f)
	}
	f, err = ParseProjection("foo{rab : bar{baz}}")
	if assert.NoError(t, err) {
		assert.Equal(t, Projection{{
			Name: "foo",
			Children: Projection{{
				Name:     "bar",
				Alias:    "rab",
				Children: Projection{{Name: "baz"}},
			}},
		}}, f)
	}
	f, err = ParseProjection(`foo(bar:"baz")`)
	if assert.NoError(t, err) {
		assert.Equal(t, Projection{{Name: "foo", Params: map[string]interface{}{"bar": "baz"}}}, f)
	}
	f, err = ParseProjection(`foo(bar:"baz\"zab")`)
	if assert.NoError(t, err) {
		assert.Equal(t, Projection{{Name: "foo", Params: map[string]interface{}{"bar": "baz\"zab"}}}, f)
	}
	f, err = ParseProjection("foo(bar:-0.2)")
	if assert.NoError(t, err) {
		assert.Equal(t, Projection{{Name: "foo", Params: map[string]interface{}{"bar": -0.2}}}, f)
	}
	f, err = ParseProjection(`foo(bar : -0.2 , baz = "zab")`)
	if assert.NoError(t, err) {
		assert.Equal(t, Projection{{Name: "foo", Params: map[string]interface{}{"bar": -0.2, "baz": "zab"}}}, f)
	}
}

func TestParseProjectionExpressionInvalid(t *testing.T) {
	_, err := ParseProjection("foo{bar,baz")
	assert.EqualError(t, err, "looking for `}' at char 11")
	_, err = ParseProjection("bar{baz}:foo")
	assert.EqualError(t, err, "invalid char `:` at 8")
	_, err = ParseProjection("foo:{bar}")
	assert.EqualError(t, err, "looking for field name at char 4")
	_, err = ParseProjection("foo{}")
	assert.EqualError(t, err, "looking for field name at char 4")
	_, err = ParseProjection("{foo}")
	assert.EqualError(t, err, "looking for field name at char 0")
	_, err = ParseProjection(",foo")
	assert.EqualError(t, err, "looking for field name at char 0")
	_, err = ParseProjection("f oo")
	assert.EqualError(t, err, "invalid char `o` at 2")
	_, err = ParseProjection("foo}")
	assert.EqualError(t, err, "looking for field name and got `}' at char 3")
	_, err = ParseProjection("foo()")
	assert.EqualError(t, err, "looking for parameter name at char 4")
	_, err = ParseProjection("foo(bar baz)")
	assert.EqualError(t, err, "looking for : at char 8")
	_, err = ParseProjection("foo(bar")
	assert.EqualError(t, err, "looking for : at char 7")
	_, err = ParseProjection(`foo(bar:"baz)`)
	assert.EqualError(t, err, "looking for \" at char 13")
	_, err = ParseProjection("foo(bar:0a)")
	assert.EqualError(t, err, "looking for `,' or ')' at char 9")
	_, err = ParseProjection("foo(bar:@toto)")
	assert.EqualError(t, err, "looking for value at char 8")
	_, err = ParseProjection("foo,")
	assert.EqualError(t, err, "looking for field name at char 4")
	_, err = ParseProjection("foo{bar,}")
	assert.EqualError(t, err, "looking for field name at char 8")

}
