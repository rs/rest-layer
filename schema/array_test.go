package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayValidatorCompile(t *testing.T) {
	v := &Array{ValuesValidator: &String{}}
	err := v.Compile()
	assert.NoError(t, err)
	v = &Array{ValuesValidator: &String{Regexp: "[invalid re"}}
	err = v.Compile()
	assert.EqualError(t, err, "invalid regexp: error parsing regexp: missing closing ]: `[invalid re`")

}

func TestArrayValidator(t *testing.T) {
	v, err := Array{ValuesValidator: &Bool{}}.Validate([]interface{}{true, false})
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{true, false}, v)
	v, err = Array{ValuesValidator: &Bool{}}.Validate([]interface{}{true, "value"})
	assert.EqualError(t, err, "invalid value at #2: not a Boolean")
	assert.Equal(t, nil, v)
	v, err = Array{ValuesValidator: &String{}}.Validate("value")
	assert.EqualError(t, err, "not an array")
	assert.Equal(t, nil, v)
}
