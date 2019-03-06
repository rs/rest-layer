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
	return v.Values.Compile(rc)
}

func (v Array) validateValues(values []interface{}, query bool) ([]interface{}, error) {
	if v.Values.Validator == nil {
		return values, nil
	}

	var vFunc func(val interface{}) (interface{}, error)
	if qv, ok := v.Values.Validator.(FieldQueryValidator); ok && query {
		vFunc = qv.ValidateQuery
	} else {
		vFunc = v.Values.Validator.Validate
	}

	for i, val := range values {
		val, err := vFunc(val)
		if err != nil {
			return nil, fmt.Errorf("invalid value at #%d: %s", i+1, err)
		}
		values[i] = val
	}
	return values, nil
}

// ValidateQuery implements FieldQueryValidator.
func (v Array) ValidateQuery(value interface{}) (interface{}, error) {
	values, isArray := value.([]interface{})
	if !isArray {
		values = append(values, value)
	}

	arr, err := v.validateValues(values, true)
	if err != nil {
		return nil, err
	}

	if !isArray {
		return arr[0], nil
	}
	return arr, nil
}

// Validate implements FieldValidator.
func (v Array) Validate(value interface{}) (interface{}, error) {
	values, ok := value.([]interface{})
	if !ok {
		return nil, errors.New("not an array")
	}
	l := len(values)
	if l < v.MinLen {
		return nil, fmt.Errorf("has fewer items than %d", v.MinLen)
	}
	if v.MaxLen > 0 && l > v.MaxLen {
		return nil, fmt.Errorf("has more items than %d", v.MaxLen)
	}
	arr, err := v.validateValues(values, false)
	if err != nil {
		return nil, err
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
