package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllOfValidatorCompile(t *testing.T) {
	v := &AllOf{&String{}}
	err := v.Compile()
	assert.NoError(t, err)
	v = &AllOf{&String{Regexp: "[invalid re"}}
	err = v.Compile()
	assert.EqualError(t, err, "invalid regexp: error parsing regexp: missing closing ]: `[invalid re`")

}

func TestAllOfValidator(t *testing.T) {
	v, err := AllOf{&Bool{}, &Bool{}}.Validate(true)
	assert.NoError(t, err)
	assert.Equal(t, true, v)
	v, err = AllOf{&Bool{}, &Bool{}}.Validate("")
	assert.EqualError(t, err, "not a Boolean")
	assert.Equal(t, nil, v)
	v, err = AllOf{&Bool{}, &String{}}.Validate(true)
	assert.EqualError(t, err, "not a string")
	assert.Equal(t, nil, v)
}
