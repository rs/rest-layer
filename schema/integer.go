package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Integer validates integer based values.
type Integer struct {
	Allowed    []int
	Boundaries *Boundaries
}

// Validate validates and normalize integer based value.
func (v Integer) Validate(value interface{}) (interface{}, error) {
	var i int
	switch val := value.(type) {
	case json.Number:
		if strings.Index(val.String(), ".") != -1 {
			return nil, errors.New("found float, integer expected")
		}
		d, _ := val.Int64()
		i = int(d)
	case int:
		i = val
	default:
		return nil, errors.New("not an integer")
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
