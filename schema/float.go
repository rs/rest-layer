package schema

import (
	"errors"
	"fmt"
)

// Boundaries defines min/max for an integer.
type Boundaries struct {
	Min float64
	Max float64
}

// Float validates float based values.
type Float struct {
	Allowed    []float64
	Boundaries *Boundaries
}

// ValidateQuery implements schema.FieldQueryValidator interface
func (v Float) ValidateQuery(value interface{}) (interface{}, error) {
	return v.parse(value)
}

func (v Float) get(value interface{}) (float64, error) {
	f, ok := value.(float64)
	if !ok {
		return 0, errors.New("not a float")
	}
	return f, nil
}

// Validate validates and normalize float based value.
func (v Float) Validate(value interface{}) (interface{}, error) {
	f, err := v.get(value)
	if err != nil {
		return nil, err
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
			// TODO: build the list of allowed values.
			return nil, fmt.Errorf("not one of the allowed values")
		}
	}
	return f, nil
}

func (v Float) parse(value interface{}) (interface{}, error) {
	f, ok := value.(float64)
	if !ok {
		return nil, errors.New("not a float")
	}
	return f, nil
}

// LessFunc implements the FieldComparator interface.
func (v Float) LessFunc() LessFunc {
	return v.less
}

func (v Float) less(value, other interface{}) bool {
	t, err1 := v.get(value)
	o, err2 := v.get(other)
	if err1 != nil || err2 != nil {
		return false
	}
	return t < o
}
