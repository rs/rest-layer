package schema

import "errors"

// AnyOf validates if any of the sub field validators validates. If any of the
// sub field validators implements the FieldSerializer interface, the *first*
// implementation which does not error will be used.
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

// Serialize attempts to serialize the value using the first available
// FieldSerializer which does not return an error. If no appropriate serializer
// is found, the input value is returned.
func (v AnyOf) Serialize(value interface{}) (interface{}, error) {
	for _, serializer := range v {
		s, ok := serializer.(FieldSerializer)
		if !ok {
			continue
		}

		v, err := s.Serialize(value)
		if err != nil {
			continue
		}
		return v, nil
	}

	return value, nil
}
