package schema

import "errors"

// AnyOf validates if any of the sub field validators validates.
type AnyOf []FieldValidator

// Compile implements the Compiler interface.
func (v AnyOf) Compile(rc ReferenceChecker) error {
	for _, sv := range v {
		if c, ok := sv.(Compiler); ok {
			if err := c.Compile(rc); err != nil {
				return err
			}
		}

	}
	return nil
}

// Validate ensures that at least one sub-validator validates.
func (v AnyOf) Validate(value interface{}) (interface{}, error) {
	for _, validator := range v {
		if value, err := validator.Validate(value); err == nil {
			return value, nil
		}
	}
	// TODO: combine errors.
	return nil, errors.New("invalid")
}
