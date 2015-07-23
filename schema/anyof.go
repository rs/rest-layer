package schema

import "errors"

// AnyOf validates if any of the sub field validators validates
type AnyOf []FieldValidator

// Validate ensures that at least one sub-validator validates
func (v AnyOf) Validate(value interface{}) (interface{}, error) {
	for _, validator := range v {
		var err error
		if value, err = validator.Validate(value); err == nil {
			return value, nil
		}
	}
	return nil, errors.New("invalid")
}
