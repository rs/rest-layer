package schema

import (
	"errors"
	"fmt"
)

// Array validates array values
type Array struct {
	// ValuesValidator is the validator to apply on array items
	ValuesValidator FieldValidator
}

// Compile implements Compiler interface
func (v *Array) Compile() (err error) {
	if c, ok := v.ValuesValidator.(Compiler); ok {
		if err = c.Compile(); err != nil {
			return
		}
	}
	return
}

// Validate implements FieldValidator
func (v Array) Validate(value interface{}) (interface{}, error) {
	arr, ok := value.([]interface{})
	if !ok {
		return nil, errors.New("not an array")
	}
	for i, val := range arr {
		if v.ValuesValidator != nil {
			val, err := v.ValuesValidator.Validate(val)
			if err != nil {
				return nil, fmt.Errorf("invalid value at #%d: %s", i+1, err)
			}
			arr[i] = val
		}
	}
	return arr, nil
}
