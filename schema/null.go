package schema

import "errors"

// Null validates that the value is null.
type Null []FieldValidator

// Validate ensures that value is null.
func (v Null) Validate(value interface{}) (interface{}, error) {
	if value != nil {
		return nil, errors.New("not null")
	}
	return value, nil
}
