package query

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/rest-layer/schema"
)

type Sort []SortField

type SortField struct {
	// Name is the name of the field to sort on.
	Name string

	// Reversed instruct to reverse the sorting if set to true.
	Reversed bool
}

// MustParseSort parses a sort expression and panics in case of error.
func MustParseSort(sort string) Sort {
	s, err := ParseSort(sort)
	if err != nil {
		panic(fmt.Sprintf("query: ParseSort(%q): %v", sort, err))
	}
	return s
}

// ParseSort parses a sort expression. A sort expression is a list of fields
// separated by comas. A field sort is reverse if preceded by a minus sign (-).
func ParseSort(sort string) (Sort, error) {
	s := Sort{}
	if strings.Trim(sort, " ") == "" {
		return s, nil
	}
	for _, f := range strings.Split(sort, ",") {
		sf := SortField{Name: strings.Trim(f, " ")}
		if sf.Name == "" || sf.Name == "-" {
			return nil, errors.New("empty sort field")
		}
		// If the field start with - (to indicate descended sort), shift it
		// before validator lookup.
		if sf.Name[0] == '-' {
			sf.Name = sf.Name[1:]
			sf.Reversed = true
		}
		s = append(s, sf)
	}
	return s, nil
}

// Validate validates the sort againast the provided validator.
func (s Sort) Validate(validator schema.Validator) error {
	for _, sf := range s {
		// Make sure the field exists.
		field := validator.GetField(sf.Name)
		if field == nil {
			return fmt.Errorf("invalid sort field: %s", sf.Name)
		}
		if !field.Sortable {
			return fmt.Errorf("field is not sortable: %s", sf.Name)
		}
	}
	return nil
}
