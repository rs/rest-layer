package schema

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type uncompilableValidator struct{}

func (v uncompilableValidator) Compile() error {
	return errors.New("compilation failed")
}

func (v uncompilableValidator) Validate(value interface{}) (interface{}, error) {
	return value, nil
}

func TestInvalidObjectValidatorCompile(t *testing.T) {
	v := &Object{}
	err := v.Compile()
	assert.Error(t, err)
}

func TestObjectValidatorCompile(t *testing.T) {
	v := &Object{
		Schema: &Schema{},
	}
	err := v.Compile()
	assert.NoError(t, err)
}

func TestObjectWithSchemaValidatorCompile(t *testing.T) {
	v := &Object{
		Schema: &Schema{
			Fields: Fields{
				"test": Field{
					Validator: &String{},
				},
			},
		},
	}
	err := v.Compile()
	assert.NoError(t, err)
}

func TestObjectWithSchemaValidatorCompileError(t *testing.T) {
	v := &Object{
		Schema: &Schema{
			Fields: Fields{
				"foo": Field{
					Validator: &uncompilableValidator{},
				},
			},
		},
	}
	err := v.Compile()
	assert.EqualError(t, err, "foo: compilation failed")
}

func TestObjectValidator(t *testing.T) {
	obj := make(map[string]interface{})
	obj["test"] = "hello"
	v := &Object{
		Schema: &Schema{
			Fields: Fields{
				"test": Field{
					Validator: &String{},
				},
			},
		},
	}
	assert.NoError(t, v.Compile())
	doc, err := v.Validate(obj)
	assert.NoError(t, err)
	assert.Equal(t, obj, doc)
}

func TestInvalidObjectValidator(t *testing.T) {
	obj := make(map[string]interface{})
	obj["test"] = 1
	v := &Object{
		Schema: &Schema{
			Fields: Fields{
				"test": Field{
					Validator: &String{},
				},
			},
		},
	}
	assert.NoError(t, v.Compile())
	_, err := v.Validate(obj)
	assert.Error(t, err)
}

func TestErrorObjectCast(t *testing.T) {
	obj := make(map[string]interface{})
	obj["test"] = 1
	v := &Object{
		Schema: &Schema{
			Fields: Fields{
				"test": Field{
					Validator: &String{},
				},
			},
		},
	}
	assert.NoError(t, v.Compile())
	_, err := v.Validate(obj)
	switch errMap := err.(type) {
	case ErrorMap:
		assert.True(t, true)
		assert.Len(t, errMap, 1)
	default:
		assert.True(t, false)
	}
}

func TestArrayOfObject(t *testing.T) {
	obj := make(map[string]interface{})
	obj["test"] = "a"
	objb := make(map[string]interface{})
	objb["test"] = "b"
	value := &Object{
		Schema: &Schema{
			Fields: Fields{
				"test": Field{
					Validator: &String{},
				},
			},
		},
	}
	array := Array{ValuesValidator: value}
	a := []interface{}{obj, objb}
	assert.NoError(t, array.Compile())
	_, err := array.Validate(a)
	assert.NoError(t, err)
}

func TestErrorArrayOfObject(t *testing.T) {
	obj := make(map[string]interface{})
	obj["test"] = "a"
	objb := make(map[string]interface{})
	objb["test"] = 1
	value := &Object{
		Schema: &Schema{
			Fields: Fields{
				"test": Field{
					Validator: &String{},
				},
			},
		},
	}
	array := Array{ValuesValidator: value}
	a := []interface{}{obj, objb}
	assert.NoError(t, array.Compile())
	_, err := array.Validate(a)
	assert.Error(t, err)
}

func TestErrorBasicMessage(t *testing.T) {
	obj := make(map[string]interface{})
	obj["test"] = 1
	v := &Object{
		Schema: &Schema{
			Fields: Fields{
				"test": Field{
					Validator: &String{},
				},
			},
		},
	}
	assert.NoError(t, v.Compile())
	_, err := v.Validate(obj)
	errMap, ok := err.(ErrorMap)
	assert.True(t, ok)
	assert.Len(t, errMap, 1)
	assert.Equal(t, "test is [not a string]", errMap.Error())
}

func Test2ErrorFieldMessages(t *testing.T) {
	obj := make(map[string]interface{})
	obj["test"] = 1
	obj["count"] = "blah"
	v := &Object{
		Schema: &Schema{
			Fields: Fields{
				"test": Field{
					Validator: &String{},
				},
				"count": Field{
					Validator: &Integer{},
				},
			},
		},
	}
	assert.NoError(t, v.Compile())
	_, err := v.Validate(obj)
	errMap, ok := err.(ErrorMap)
	assert.True(t, ok)
	assert.Len(t, errMap, 2)
	assert.Equal(t, "count is [not an integer], test is [not a string]", errMap.Error())
}

func TestErrorMessagesForObjectValidatorEmbeddedInArray(t *testing.T) {
	obj := make(map[string]interface{})
	obj["test"] = 1
	obj["isUp"] = "false"
	value := &Object{
		Schema: &Schema{
			Fields: Fields{
				"test": Field{
					Validator: &String{},
				},
				"isUp": Field{
					Validator: &Bool{},
				},
			},
		},
	}
	assert.NoError(t, value.Compile())

	array := Array{ValuesValidator: value}

	// Not testing multiple array values being errors because Array
	// implementation stops validating on first error found in array.
	a := []interface{}{obj}
	_, err := array.Validate(a)
	assert.Error(t, err)
	assert.Equal(t, "invalid value at #1: isUp is [not a Boolean], test is [not a string]", err.Error())
}
