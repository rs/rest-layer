package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/rs/rest-layer/schema"
)

// Lookup holds key/value pairs used to select items on a resource
type Lookup struct {
	// Fields are field=>value pairs that must be equal. The main difference with the
	// Filter field is that Fields are set by the route while Filter is defined by the
	// "filter" parameter.
	Fields map[string]interface{}
	// The client supplied filter. Filter is a MongoDB inspired query with a more limited
	// set of capabilities. See [README](https://github.com/rs/rest-layer#filtering)
	// for more info.
	Filter map[string]interface{}
	// The client supplied soft. Sort is a list of resource fields or sub-fields separated
	// by comas (,). To invert the sort, a minus (-) can be prefixed.
	// See [README](https://github.com/rs/rest-layer#sorting) for more info.
	Sort []string
}

// NewLookup creates an empty lookup object
func NewLookup() *Lookup {
	return &Lookup{
		Fields: map[string]interface{}{},
		Filter: map[string]interface{}{},
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
		if field := validator.GetField(f[i:]); field == nil {
			return fmt.Errorf("invalid sort field: %s", f[i:])
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
	var j interface{}
	json.Unmarshal([]byte(filter), &j)
	f, ok := j.(map[string]interface{})
	if !ok {
		return errors.New("must be a JSON object")
	}
	if err := validateFilter(f, validator, ""); err != nil {
		return err
	}
	l.Filter = f
	return nil
}

// validateFilter recursively validates the format of a filter query
func validateFilter(f map[string]interface{}, validator schema.Validator, parentKey string) error {
	for key, exp := range f {
		switch key {
		case "$ne":
			op := key
			if parentKey == "" {
				return fmt.Errorf("%s can't be at first level", op)
			}
			if field := validator.GetField(parentKey); field != nil {
				if field.Validator != nil {
					if _, err := field.Validator.Validate(exp); err != nil {
						return fmt.Errorf("invalid filter expression for field `%s': %s", parentKey, err)
					}
				}
			} else {
				return fmt.Errorf("unknown filter field: %s", parentKey)
			}
		case "$gt", "$gte", "$lt", "$lte":
			op := key
			if parentKey == "" {
				return fmt.Errorf("%s can't be at first level", op)
			}
			if _, ok := isNumber(exp); !ok {
				return fmt.Errorf("%s: value for %s must be a number", parentKey, op)
			}
			if field := validator.GetField(parentKey); field != nil {
				if field.Validator != nil {
					switch field.Validator.(type) {
					case schema.Integer, schema.Float:
						if _, err := field.Validator.Validate(exp); err != nil {
							return fmt.Errorf("invalid filter expression for field `%s': %s", parentKey, err)
						}
					default:
						return fmt.Errorf("%s: cannot apply %s operation on a non numerical field", parentKey, op)
					}
				}
			} else {
				return fmt.Errorf("unknown filter field: %s", parentKey)
			}
		case "$in", "$nin":
			op := key
			if parentKey == "" {
				return fmt.Errorf("%s can't be at first level", op)
			}
			if _, ok := exp.(map[string]interface{}); ok {
				return fmt.Errorf("%s: value for %s can't be a dict", parentKey, op)
			}
			if field := validator.GetField(parentKey); field != nil {
				if field.Validator != nil {
					values, ok := exp.([]interface{})
					if !ok {
						values = []interface{}{exp}
					}
					for _, value := range values {
						if _, err := field.Validator.Validate(value); err != nil {
							return fmt.Errorf("invalid filter expression (%s) for field `%s': %s", value, parentKey, err)
						}
					}
				}
			} else {
				return fmt.Errorf("unknown filter field: %s", parentKey)
			}
		case "$or":
			var subFilters []interface{}
			var ok bool
			if subFilters, ok = exp.([]interface{}); !ok {
				return errors.New("value for $or must be an array of dicts")
			}
			if len(subFilters) < 2 {
				return errors.New("$or must contain at least to elements")
			}
			for _, subFilter := range subFilters {
				if sf, ok := subFilter.(map[string]interface{}); !ok {
					return errors.New("value for $or must be an array of dicts")
				} else if err := validateFilter(sf, validator, ""); err != nil {
					return err
				}
			}
		default:
			// Exact match
			if parentKey != "" {
				return fmt.Errorf("%s: invalid expression", parentKey)
			}
			if subFilter, ok := exp.(map[string]interface{}); ok {
				if err := validateFilter(subFilter, validator, key); err != nil {
					return err
				}
			} else {
				if field := validator.GetField(key); field != nil {
					if field.Validator != nil {
						if _, err := field.Validator.Validate(exp); err != nil {
							return fmt.Errorf("invalid filter expression for field `%s': %s", key, err)
						}
					}
				} else {
					return fmt.Errorf("unknown filter field: %s", key)
				}
			}
		}
	}
	return nil
}

// Match evaluates lookup's fields and filter on the provided payload
// and tells if it match
func (l *Lookup) Match(payload map[string]interface{}) bool {
	for field, value := range l.Fields {
		if !reflect.DeepEqual(payload[field], value) {
			return false
		}
	}
	return matchFilter(l.Filter, payload, "")
}

func matchFilter(f map[string]interface{}, payload map[string]interface{}, parentKey string) bool {
	for key, exp := range f {
		switch key {
		case "$ne":
			if reflect.DeepEqual(getField(payload, parentKey), exp) {
				return false
			}
		case "$gt":
			n1, ok1 := isNumber(exp)
			n2, ok2 := isNumber(getField(payload, parentKey))
			if !(ok1 && ok2 && (n1 < n2)) {
				return false
			}
		case "$gte":
			n1, ok1 := isNumber(exp)
			n2, ok2 := isNumber(getField(payload, parentKey))
			if !(ok1 && ok2 && (n1 <= n2)) {
				return false
			}
		case "$lt":
			n1, ok1 := isNumber(exp)
			n2, ok2 := isNumber(getField(payload, parentKey))
			if !(ok1 && ok2 && (n1 > n2)) {
				return false
			}
		case "$lte":
			n1, ok1 := isNumber(exp)
			n2, ok2 := isNumber(getField(payload, parentKey))
			if !(ok1 && ok2 && (n1 >= n2)) {
				return false
			}
		case "$in":
			if !isIn(exp, getField(payload, parentKey)) {
				return false
			}
		case "$nin":
			if isIn(exp, getField(payload, parentKey)) {
				return false
			}
		case "$or":
			pass := false
			if subFilters, ok := exp.([]interface{}); ok {
				// Run each sub filters like a root filter, stop/pass on first match
				for _, subFilter := range subFilters {
					if matchFilter(subFilter.(map[string]interface{}), payload, "") {
						pass = true
						break
					}
				}
			}
			if !pass {
				return false
			}
		default:
			// Exact match
			if subFilter, ok := exp.(map[string]interface{}); ok {
				if !matchFilter(subFilter, payload, key) {
					return false
				}
			} else if !reflect.DeepEqual(getField(payload, key), exp) {
				return false
			}
		}
	}
	return true
}

// applyFields appends lookup fields to a payload
func (l *Lookup) applyFields(payload map[string]interface{}) {
	for field, value := range l.Fields {
		payload[field] = value
	}
}
