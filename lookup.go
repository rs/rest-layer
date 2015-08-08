package rest

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/rest-layer/schema"
)

// Lookup holds filter and sort used to select items in a resource collection
type Lookup interface {
	// The client supplied filter. Filter is a MongoDB inspired query with a more limited
	// set of capabilities. See [README](https://github.com/rs/rest-layer#filtering)
	// for more info.
	Filter() schema.Query
	// The client supplied soft. Sort is a list of resource fields or sub-fields separated
	// by comas (,). To invert the sort, a minus (-) can be prefixed.
	// See [README](https://github.com/rs/rest-layer#sorting) for more info.
	Sort() []string
}

type lookup struct {
	// The client supplied filter. Filter is a MongoDB inspired query with a more limited
	// set of capabilities. See [README](https://github.com/rs/rest-layer#filtering)
	// for more info.
	filter schema.Query
	// The client supplied soft. Sort is a list of resource fields or sub-fields separated
	// by comas (,). To invert the sort, a minus (-) can be prefixed.
	// See [README](https://github.com/rs/rest-layer#sorting) for more info.
	sort []string
}

// newLookup creates an empty lookup object
func newLookup() *lookup {
	return &lookup{
		filter: schema.Query{},
		sort:   []string{},
	}
}

func (l *lookup) Sort() []string {
	return l.sort
}

func (l *lookup) Filter() schema.Query {
	return l.filter
}

// setSort parses and validate a sort parameter and set it as lookup's Sort
func (l *lookup) setSort(sort string, validator schema.Validator) error {
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
	l.sort = sorts
	return nil
}

// addFilter parses and validate a filter parameter and add it to lookup's filter
//
// The filter query is validated against the provided validator to ensure all queried
// fields exists and are of the right type.
func (l *lookup) addFilter(filter string, validator schema.Validator) error {
	f, err := schema.ParseQuery(filter, validator)
	if err != nil {
		return err
	}
	l.addQuery(f)
	return nil
}

func (l *lookup) addQuery(query schema.Query) {
	if l.filter == nil {
		l.filter = query
		return
	}
	for _, exp := range query {
		l.filter = append(l.filter, exp)
	}
}
