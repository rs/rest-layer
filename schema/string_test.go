package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringValidator(t *testing.T) {
	s, err := String{}.Validate("foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", s)
	s, err = String{MaxLen: 2}.Validate("foo")
	assert.EqualError(t, err, "is longer than 2")
	assert.Nil(t, s)
	s, err = String{MaxLen: 4}.Validate("foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", s)
	s, err = String{MinLen: 4}.Validate("foo")
	assert.EqualError(t, err, "is shorter than 4")
	assert.Nil(t, s)
	s, err = String{MinLen: 2}.Validate("foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", s)
	s, err = String{Allowed: []string{"foo", "bar"}}.Validate("foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", s)
	s, err = String{Allowed: []string{"bar", "baz"}}.Validate("foo")
	assert.EqualError(t, err, "not one of [bar, baz]")
	assert.Nil(t, s)
	v := String{Regexp: "^f.o$"}
	assert.NoError(t, v.Compile(nil))
	s, err = v.Validate("foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", s)
	v = String{Regexp: "^bar$"}
	assert.NoError(t, v.Compile(nil))
	s, err = v.Validate("foo")
	assert.EqualError(t, err, "does not match ^bar$")
	assert.Nil(t, s)
	v = String{Regexp: "^bar["}
	assert.EqualError(t, v.Compile(nil), "invalid regexp: error parsing regexp: missing closing ]: `[`")
}
