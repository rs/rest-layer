package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDictValidatorCompile(t *testing.T) {
	v := &Dict{KeysValidator: &String{}, ValuesValidator: &String{}}
	err := v.Compile()
	assert.NoError(t, err)
	v = &Dict{
		KeysValidator: &String{Regexp: "[invalid re"},
	}
	err = v.Compile()
	assert.EqualError(t, err, "invalid regexp: error parsing regexp: missing closing ]: `[invalid re`")
	v = &Dict{
		ValuesValidator: &String{Regexp: "[invalid re"},
	}
	err = v.Compile()
	assert.EqualError(t, err, "invalid regexp: error parsing regexp: missing closing ]: `[invalid re`")
}

func TestDictValidator(t *testing.T) {
	v, err := Dict{KeysValidator: &String{}}.Validate(map[string]interface{}{"foo": true, "bar": false})
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"foo": true, "bar": false}, v)
	v, err = Dict{KeysValidator: &String{MinLen: 3}}.Validate(map[string]interface{}{"foo": true, "ba": false})
	assert.EqualError(t, err, "invalid key `ba': is shorter than 3")
	assert.Equal(t, nil, v)
	v, err = Dict{ValuesValidator: &Bool{}}.Validate(map[string]interface{}{"foo": true, "bar": false})
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"foo": true, "bar": false}, v)
	v, err = Dict{ValuesValidator: &Bool{}}.Validate(map[string]interface{}{"foo": true, "bar": "value"})
	assert.EqualError(t, err, "invalid value for key `bar': not a Boolean")
	assert.Equal(t, nil, v)
	v, err = Dict{ValuesValidator: &String{}}.Validate("value")
	assert.EqualError(t, err, "not a dict")
	assert.Equal(t, nil, v)
	v, err = Dict{ValuesValidator: &String{}}.Validate([]interface{}{"value"})
	assert.EqualError(t, err, "not a dict")
	assert.Equal(t, nil, v)

}
