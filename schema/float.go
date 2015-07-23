package schema

import (
	"errors"
	"fmt"
)

// Float validates float based values
type Float struct {
	Allowed []float64
	Min     *float64
	Max     *float64
}

// Validate validates and normalize float based value
func (v Float) Validate(value interface{}) (interface{}, error) {
	f, ok := value.(float64)
	if !ok {
		return nil, errors.New("not a float")
	}
	if v.Min != nil && f < *v.Min {
		return nil, fmt.Errorf("is lower than %f", *v.Min)
	}
	if v.Max != nil && f > *v.Max {
		return nil, fmt.Errorf("is greater than %f", *v.Max)
	}
	if len(v.Allowed) > 0 {
		found := false
		for _, allowed := range v.Allowed {
			if f == allowed {
				found = true
				break
			}
		}
		if !found {
			// TODO: build the list of allowed values
			return nil, fmt.Errorf("not one of the allowed values")
		}
	}
	return f, nil
}
