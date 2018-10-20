package schema

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringQueryValidator(t *testing.T) {
	cases := []struct {
		name          string
		field         String
		input, expect interface{}
		err           error
	}{
		{`String.ValidateQuery(string)`, String{}, "foo", "foo", nil},
		{`String.ValidateQuery(string)-ouf range`, String{MaxLen: 2}, "foo", "foo", nil},
		{`String.ValidateQuery(string)-not allowed`, String{Allowed: []string{"bar", "baz"}}, "foo", "foo", nil},
		{"String.ValidateQuery(int)", String{}, 1, nil, errors.New("not a string")},
	}
	for i := range cases {
		tt := cases[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := (tt.field).ValidateQuery(tt.input)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("unexpected error for `%v`\ngot:  %v\nwant: %v", tt.input, err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("invalid output for `%v`:\ngot:  %#v\nwant: %#v", tt.input, got, tt.expect)
			}
		})
	}
}

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
	s, err = String{}.ValidateQuery(1)
	assert.EqualError(t, err, "not a string")
	assert.Nil(t, s)
}
