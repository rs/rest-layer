package schema

import (
	"errors"
	"fmt"
	"math"
)

// Integer validates integer based values.
type Integer struct {
	Allowed    []int
	Boundaries *Boundaries
}

// ValidateQuery implements schema.FieldQueryValidator interface
func (v Integer) ValidateQuery(value interface{}) (interface{}, error) {
	return v.parse(value)
}

func (v Integer) get(value interface{}) (int, error) {
	i, ok := value.(int)
	if !ok {
		return 0, errors.New("not an integer")
	}
	return i, nil
}

// Validate validates and normalize integer based value.
func (v Integer) Validate(value interface{}) (interface{}, error) {
	val, err := v.parse(value)
	if err != nil {
		return nil, err
	}
	i, err := v.get(val)
	if err != nil {
		return nil, err
	}
	if v.Boundaries != nil {
		if float64(i) < v.Boundaries.Min {
			return nil, fmt.Errorf("is lower than %.0f", v.Boundaries.Min)
		}
		if float64(i) > v.Boundaries.Max {
			return nil, fmt.Errorf("is greater than %.0f", v.Boundaries.Max)
		}
	}
	if len(v.Allowed) > 0 {
		found := false
		for _, allowed := range v.Allowed {
			if i == allowed {
				found = true
				break
			}
		}
		if !found {
			// TODO: build the list of allowed values.
			return nil, fmt.Errorf("not one of the allowed values")
		}
	}
	return i, nil
}

func (v Integer) parse(value interface{}) (interface{}, error) {
	if f, ok := value.(float64); ok {
		// JSON unmarshaling treat all numbers as float64, try to convert it to
		// int if not fraction.
		i, frac := math.Modf(f)
		if frac == 0.0 {
			v := int(i)
			value = v
		}
	}
	i, ok := value.(int)
	if !ok {
		return nil, errors.New("not an integer")
	}
	return i, nil
}

// LessFunc implements the FieldComparator interface.
func (v Integer) LessFunc() LessFunc {
	return v.less
}

func (v Integer) less(value, other interface{}) bool {
	t, err1 := v.get(value)
	o, err2 := v.get(other)
	if err1 != nil || err2 != nil {
		return false
	}
	return t < o
}
