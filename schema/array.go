package schema

import (
	"errors"
	"fmt"
	"strconv"
)

// Array validates array values.
type Array struct {
	// Values describes the properties for each array item.
	Values Field
	// MinLen defines the minimum array length (default 0).
	MinLen int
	// MaxLen defines the maximum array length (default no limit).
	MaxLen int
}

// Compile implements the ReferenceCompiler interface.
func (v *Array) Compile(rc ReferenceChecker) (err error) {
	if c, ok := v.Values.Validator.(Compiler); ok {
		if err = c.Compile(rc); err != nil {
			return
		}
	}
	return
}

// Validate implements FieldValidator.
func (v Array) Validate(value interface{}) (interface{}, error) {
	arr, ok := value.([]interface{})
	if !ok {
		return nil, errors.New("not an array")
	}
	for i, val := range arr {
		if v.Values.Validator != nil {
			val, err := v.Values.Validator.Validate(val)
			if err != nil {
				return nil, fmt.Errorf("invalid value at #%d: %s", i+1, err)
			}
			arr[i] = val
		}
	}
	l := len(arr)
	if l < v.MinLen {
		return nil, fmt.Errorf("has fewer items than %d", v.MinLen)
	}
	if v.MaxLen > 0 && l > v.MaxLen {
		return nil, fmt.Errorf("has more items than %d", v.MaxLen)
	}
	return arr, nil
}

// GetField implements the FieldGetter interface. It will return
// a Field if name corespond to a legal array index according to
// parameters set on v.
func (v Array) GetField(name string) *Field {
	if i, err := strconv.Atoi(name); err != nil {
		return nil
	} else if i < 0 || (v.MaxLen > 0 && i >= v.MaxLen) {
		return nil
	}
	return &v.Values
}
