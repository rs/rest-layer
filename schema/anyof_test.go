package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnyOfValidatorCompile(t *testing.T) {
	v := &AnyOf{&String{}}
	err := v.Compile()
	assert.NoError(t, err)
	v = &AnyOf{&String{Regexp: "[invalid re"}}
	err = v.Compile()
	assert.EqualError(t, err, "invalid regexp: error parsing regexp: missing closing ]: `[invalid re`")

}

func TestAnyOfValidator(t *testing.T) {
	v, err := AnyOf{&Bool{}, &Bool{}}.Validate(true)
	assert.NoError(t, err)
	assert.Equal(t, true, v)
	v, err = AnyOf{&Bool{}, &Bool{}}.Validate("")
	assert.EqualError(t, err, "invalid")
	assert.Equal(t, nil, v)
	v, err = AnyOf{&Bool{}, &String{}}.Validate(true)
	assert.NoError(t, err)
	assert.Equal(t, true, v)
}
