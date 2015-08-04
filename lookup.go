package rest

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/rest-layer/schema"
)

// Lookup holds key/value pairs used to select items on a resource
type Lookup struct {
	// The client supplied filter. Filter is a MongoDB inspired query with a more limited
	// set of capabilities. See [README](https://github.com/rs/rest-layer#filtering)
	// for more info.
	Filter schema.Query
	// The client supplied soft. Sort is a list of resource fields or sub-fields separated
	// by comas (,). To invert the sort, a minus (-) can be prefixed.
	// See [README](https://github.com/rs/rest-layer#sorting) for more info.
	Sort []string
}

// NewLookup creates an empty lookup object
func NewLookup() *Lookup {
	return &Lookup{
		Filter: schema.Query{},
		Sort:   []string{},
	}
}

// SetSort parses and validate a sort parameter and set it as lookup's Sort
func (l *Lookup) SetSort(sort string, validator schema.Validator) error {
	sorts := []string{}
	for _, f := range strings.Split(sort, ",") {
		f = strings.Trim(f, " ")
		if f == "" {
			return errors.New("empty soft field")
		}
		// If the field start with - (to indicate descended sort), shift it before
		// validator lookup
		i := 0
		if f[0] == '-' {
			i = 1
		}
		// Make sure the field exists
		field := validator.GetField(f[i:])
		if field == nil {
			return fmt.Errorf("invalid sort field: %s", f[i:])
		}
		if !field.Sortable {
			return fmt.Errorf("field is not sortable: %s", f[i:])
		}
		sorts = append(sorts, f)
	}
	l.Sort = sorts
	return nil
}

// SetFilter parses and validate a filter parameter and set it as lookup's Filter
//
// The filter query is validated against the provided validator to ensure all queried
// fields exists and are of the right type.
func (l *Lookup) SetFilter(filter string, validator schema.Validator) error {
	f, err := schema.ParseQuery(filter, validator)
	if err != nil {
		return err
	}
	l.Filter = f
	return nil
}

// Match evaluates lookup's filter on the provided payload and tells if it match
func (l *Lookup) Match(payload map[string]interface{}) bool {
	return l.Filter.Match(payload)
}
