package schema

import (
	"errors"
	"fmt"
	"math"
)

// Integer validates integer based values
type Integer struct {
	Allowed []int
	Min     *int
	Max     *int
}

// Validate validates and normalize integer based value
func (v Integer) Validate(value interface{}) (interface{}, error) {
	if f, ok := value.(float64); ok {
		// JSON unmarshaling treat all numbers as float64, try to convert it to int if not fraction
		i, frac := math.Modf(f)
		if frac == 0.0 {
			v := int(i)
			value = &v
		}
	}
	i, ok := value.(*int)
	if !ok {
		return nil, errors.New("not an integer")
	}
	if v.Min != nil && *i < *v.Min {
		return nil, fmt.Errorf("is lower than %d", *v.Min)
	}
	if v.Max != nil && *i > *v.Max {
		return nil, fmt.Errorf("is greater than %d", *v.Max)
	}
	if len(v.Allowed) > 0 {
		found := false
		for _, allowed := range v.Allowed {
			if *i == allowed {
				found = true
				break
			}
		}
		if !found {
			// TODO: build the list of allowed values
			return nil, fmt.Errorf("not one of the allowed values")
		}
	}
	return i, nil
}
