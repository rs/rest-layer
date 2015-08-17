package schema

import (
	"errors"
	"fmt"
)

// Boundaries defines min/max for an integer
type Boundaries struct {
	Min float64
	Max float64
}

// Float validates float based values
type Float struct {
	Allowed    []float64
	Boundaries *Boundaries
}

// Validate validates and normalize float based value
func (v Float) Validate(value interface{}) (interface{}, error) {
	f, ok := value.(float64)
	if !ok {
		return nil, errors.New("not a float")
	}
	if v.Boundaries != nil {
		if f < v.Boundaries.Min {
			return nil, fmt.Errorf("is lower than %.2f", v.Boundaries.Min)
		}
		if f > v.Boundaries.Max {
			return nil, fmt.Errorf("is greater than %.2f", v.Boundaries.Max)
		}
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
