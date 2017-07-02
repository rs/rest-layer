package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
)

// Lookup holds filter and sort used to select items in a resource collection.
type Lookup struct {
	// The client supplied filter. Filter is a MongoDB inspired query with a
	// more limited set of capabilities. See
	// https://github.com/rs/rest-layer#filtering for more info.
	filter query.Query
	// The client supplied soft. Sort is a list of resource fields or sub-fields
	// separated by comas (,). To invert the sort, a minus (-) can be prefixed.
	// See https://github.com/rs/rest-layer#sorting for more info.
	sort []string
	// The client supplied selector. Selector is a way for the client to
	// reformat the resource representation at runtime by defining which fields
	// should be included in the document. The REST Layer selector language
	// allows field aliasing, field transformation with parameters and
	// sub-item/collection embedding.
	selector []Field
}

// Field is used with Lookup.selector to reformat the resource representation at
// runtime using a field selection language inspired by GraphQL.
type Field struct {
	// Name is the name of the field as define in the resource's schema.
	Name string
	// Alias is the wanted name in the representation.
	Alias string
	// Params defines a list of params to be sent to the field's param handler
	// if any.
	Params map[string]interface{}
	// Fields holds references to child fields if any.
	Fields []Field
}

// NewLookup creates an empty lookup object
func NewLookup() *Lookup {
	return &Lookup{
		filter: query.Query{},
		sort:   []string{},
	}
}

// NewLookupWithQuery creates an empty lookup object with a given query.
func NewLookupWithQuery(q query.Query) *Lookup {
	return &Lookup{
		filter: q,
		sort:   []string{},
	}
}

// Sort is a list of resource fields or sub-fields separated by comas (,). To
// invert the sort, a minus (-) can be prefixed.
//
// See https://github.com/rs/rest-layer#sorting for more info.
func (l *Lookup) Sort() []string {
	return l.sort
}

// Filter is a MongoDB inspired query with a more limited set of capabilities.
//
// See https://github.com/rs/rest-layer#filtering for more info.
func (l *Lookup) Filter() query.Query {
	return l.filter
}

// SetSorts set the sort fields with a pre-parsed list of fields to sort on.
// This method doesn't validate sort fields.
func (l *Lookup) SetSorts(sorts []string) {
	l.sort = sorts
}

// SetSort parses and validate a sort parameter and set it as lookup's Sort
func (l *Lookup) SetSort(sort string, v schema.Validator) error {
	sorts := []string{}
	for _, f := range strings.Split(sort, ",") {
		f = strings.Trim(f, " ")
		if f == "" {
			return errors.New("empty soft field")
		}
		// If the field start with - (to indicate descended sort), shift it
		// before validator lookup.
		i := 0
		if f[0] == '-' {
			i = 1
		}
		// Make sure the field exists.
		field := v.GetField(f[i:])
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

// AddFilter parses and validate a filter parameter and add it to lookup's
// filter.
//
// The filter query is validated against the provided validator to ensure all
// queried fields exists and are of the right type.
func (l *Lookup) AddFilter(filter string, v schema.Validator) error {
	f, err := query.Parse(filter)
	if err != nil {
		return err
	}
	if err = f.Validate(v); err != nil {
		return err
	}
	l.AddQuery(f)
	return nil
}

// AddQuery add an existing query.Query to the lookup's filters
func (l *Lookup) AddQuery(query query.Query) {
	if l.filter == nil {
		l.filter = query
		return
	}
	for _, exp := range query {
		l.filter = append(l.filter, exp)
	}
}

// SetSelector parses a selector expression, validates it and assign it to the
// current Lookup.
func (l *Lookup) SetSelector(s string, v schema.Validator) error {
	pos := 0
	selector, err := parseSelectorExpression([]byte(s), &pos, len(s), false)
	if err != nil {
		return err
	}
	if err = validateSelector(selector, v); err != nil {
		return err
	}
	l.selector = selector
	return nil
}

// ReferenceResolver is a function resolving a reference to another field.
type ReferenceResolver func(path string) (*Resource, error)

// ApplySelector applies fields filtering / rename to the payload fields.
func (l *Lookup) ApplySelector(ctx context.Context, v schema.Validator, p map[string]interface{}, resolver ReferenceResolver) (map[string]interface{}, error) {
	payload, err := applySelector(ctx, l.selector, v, p, resolver)
	if err == nil {
		// The resulting payload may contain some asyncSelector, we must execute
		// them concurrently until there's no more.
		err = resolveAsyncSelectors(ctx, payload)
	}
	return payload, err
}
