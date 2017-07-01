package schema

import "errors"

// Bool validates Boolean based values.
type Bool struct {
}

// Validate validates and normalize Boolean based value.
func (v Bool) Validate(value interface{}) (interface{}, error) {
	if _, ok := value.(bool); !ok {
		return nil, errors.New("not a Boolean")
	}
	return value, nil
}
