package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
)

// Query defines an expression against a schema to perform a match on schema's
// data.
type Query []Expression

// Expression is a query or query component that can be matched against a
// payload.
type Expression interface {
	Match(payload map[string]interface{}) bool
}

// Value represents any kind of value to use in query.
type Value interface{}

// And joins query clauses with a logical AND, returns all documents
// that match the conditions of both clauses.
type And []Expression

// Or joins query clauses with a logical OR, returns all documents that
// match the conditions of either clause.
type Or []Expression

// In matches any of the values specified in an array.
type In struct {
	Field  string
	Values []Value
}

// NotIn matches none of the values specified in an array.
type NotIn struct {
	Field  string
	Values []Value
}

// Equal matches all values that are equal to a specified value.
type Equal struct {
	Field string
	Value Value
}

// NotEqual matches all values that are not equal to a specified value.
type NotEqual struct {
	Field string
	Value Value
}

// Exist matches all values which are present, even if nil.
type Exist struct {
	Field string
}

// NotExist matches all values which are absent.
type NotExist struct {
	Field string
}

// GreaterThan matches values that are greater than a specified value.
type GreaterThan struct {
	Field string
	Value float64
}

// GreaterOrEqual matches values that are greater than or equal to a specified
// value.
type GreaterOrEqual struct {
	Field string
	Value float64
}

// LowerThan matches values that are less than a specified value.
type LowerThan struct {
	Field string
	Value float64
}

// LowerOrEqual matches values that are less than or equal to a specified value.
type LowerOrEqual struct {
	Field string
	Value float64
}

// Regex matches values that match to a specified regular expression.
type Regex struct {
	Field string
	Value *regexp.Regexp
}

// NewQuery returns a new query with the provided key/value validated against
// validator.
func NewQuery(q map[string]interface{}, validator Validator) (Query, error) {
	return validateQuery(q, validator, "")
}

// ParseQuery parses and validate a query as string.
func ParseQuery(query string, validator Validator) (Query, error) {
	var j interface{}
	if err := json.Unmarshal([]byte(query), &j); err != nil {
		return nil, errors.New("must be valid JSON")
	}
	q, ok := j.(map[string]interface{})
	if !ok {
		return nil, errors.New("must be a JSON object")
	}
	return validateQuery(q, validator, "")
}

// validateQuery recursively validates and cast a query.
func validateQuery(q map[string]interface{}, validator Validator, parentKey string) (Query, error) {
	queries := Query{}
	for key, exp := range q {
		switch key {
		case "$regex":
			if parentKey == "" {
				return nil, errors.New("$regex can't be at first level")
			}
			if regex, ok := exp.(string); ok {
				v, err := regexp.Compile(regex)
				if err != nil {
					return nil, fmt.Errorf("$regex: invalid regex: %v", err)
				}
				queries = append(queries, Regex{Field: parentKey, Value: v})
			}
		case "$exists":
			if parentKey == "" {
				return nil, errors.New("$exists can't be at first level")
			}
			positive, ok := exp.(bool)
			if !ok {
				return nil, errors.New("$exists can only get Boolean as value")
			}
			if positive {
				queries = append(queries, Exist{Field: parentKey})
			} else {
				queries = append(queries, NotExist{Field: parentKey})
			}
		case "$ne":
			op := key
			if parentKey == "" {
				return nil, fmt.Errorf("%s can't be at first level", op)
			}
			if field := validator.GetField(parentKey); field != nil {
				if field.Validator != nil {
					if _, err := field.Validator.Validate(exp); err != nil {
						return nil, fmt.Errorf("invalid query expression for field `%s': %s", parentKey, err)
					}
				}
			}
			queries = append(queries, NotEqual{Field: parentKey, Value: exp})
		case "$gt", "$gte", "$lt", "$lte":
			op := key
			if parentKey == "" {
				return nil, fmt.Errorf("%s can't be at first level", op)
			}
			n, ok := isNumber(exp)
			if !ok {
				return nil, fmt.Errorf("%s: value for %s must be a number", parentKey, op)
			}
			if field := validator.GetField(parentKey); field != nil {
				if field.Validator != nil {
					switch field.Validator.(type) {
					case *Integer, *Float, Integer, Float:
						if _, err := field.Validator.Validate(exp); err != nil {
							return nil, fmt.Errorf("invalid query expression for field `%s': %s", parentKey, err)
						}
					default:
						return nil, fmt.Errorf("%s: cannot apply %s operation on a non numerical field", parentKey, op)
					}
				}
			}
			switch op {
			case "$gt":
				queries = append(queries, GreaterThan{Field: parentKey, Value: n})
			case "$gte":
				queries = append(queries, GreaterOrEqual{Field: parentKey, Value: n})
			case "$lt":
				queries = append(queries, LowerThan{Field: parentKey, Value: n})
			case "$lte":
				queries = append(queries, LowerOrEqual{Field: parentKey, Value: n})
			}
		case "$in", "$nin":
			op := key
			if parentKey == "" {
				return nil, fmt.Errorf("%s can't be at first level", op)
			}
			if _, ok := exp.(map[string]interface{}); ok {
				return nil, fmt.Errorf("%s: value for %s can't be a dict", parentKey, op)
			}
			values := []Value{}
			if field := validator.GetField(parentKey); field != nil {
				vals, ok := exp.([]interface{})
				if !ok {
					vals = []interface{}{exp}
				}
				if field.Validator != nil {
					for _, v := range vals {
						if _, err := field.Validator.Validate(v); err != nil {
							return nil, fmt.Errorf("invalid query expression (%s) for field `%s': %s", v, parentKey, err)
						}
					}
				}
				for _, v := range vals {
					values = append(values, v)
				}
			}
			switch op {
			case "$in":
				queries = append(queries, In{Field: parentKey, Values: values})
			case "$nin":
				queries = append(queries, NotIn{Field: parentKey, Values: values})
			}
		case "$or", "$and":
			op := key
			var subQueries []interface{}
			var ok bool
			if subQueries, ok = exp.([]interface{}); !ok {
				return nil, fmt.Errorf("value for %s must be an array of dicts", op)
			}
			if len(subQueries) < 2 {
				return nil, fmt.Errorf("%s must contain at least to elements", op)
			}
			// Cast map to Query object
			castedExp := []Expression{}
			for _, subQuery := range subQueries {
				sq, ok := subQuery.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("value for %s must be an array of dicts", op)
				}
				query, err := validateQuery(sq, validator, "")
				if err != nil {
					return nil, err
				}
				castedExp = append(castedExp, query...)
			}
			switch op {
			case "$or":
				queries = append(queries, Or(castedExp))
			case "$and":
				queries = append(queries, And(castedExp))
			}
		default:
			// Field query
			field := validator.GetField(key)
			if field == nil {
				return nil, fmt.Errorf("unknown query field: %s", key)
			}
			if !field.Filterable {
				return nil, fmt.Errorf("field is not filterable: %s", key)
			}
			if parentKey != "" {
				return nil, fmt.Errorf("%s: invalid expression", parentKey)
			}
			if subQuery, ok := exp.(map[string]interface{}); ok {
				sq, err := validateQuery(subQuery, validator, key)
				if err != nil {
					return nil, err
				}
				queries = append(queries, sq...)
			} else {
				// Exact match
				if field.Validator != nil {
					if _, err := field.Validator.Validate(exp); err != nil {
						return nil, fmt.Errorf("invalid query expression for field `%s': %s", key, err)
					}
				}
				queries = append(queries, Equal{Field: key, Value: exp})
			}
		}
	}
	return queries, nil
}

// Match implements Expression interface.
func (e Query) Match(payload map[string]interface{}) bool {
	// Run each sub queries like a root query, stop/pass on first match.
	for _, subQuery := range e {
		if !subQuery.Match(payload) {
			return false
		}
	}
	return true
}

// Match implements Expression interface.
func (e And) Match(payload map[string]interface{}) bool {
	// Run each sub queries like a root query, stop/pass on first match.
	for _, subQuery := range e {
		if !subQuery.Match(payload) {
			return false
		}
	}
	return true
}

// Match implements Expression interface
func (e Or) Match(payload map[string]interface{}) bool {
	// Run each sub queries like a root query, stop/pass on first match.
	for _, subQuery := range e {
		if subQuery.Match(payload) {
			return true
		}
	}
	return false
}

// Match implements Expression interface.
func (e In) Match(payload map[string]interface{}) bool {
	value := getField(payload, e.Field)
	for _, v := range e.Values {
		if reflect.DeepEqual(v, value) {
			return true
		}
	}
	return false
}

// Match implements Expression interface.
func (e NotIn) Match(payload map[string]interface{}) bool {
	value := getField(payload, e.Field)
	for _, v := range e.Values {
		if reflect.DeepEqual(v, value) {
			return false
		}
	}
	return true
}

// Match implements Expression interface
func (e Equal) Match(payload map[string]interface{}) bool {
	return reflect.DeepEqual(getField(payload, e.Field), e.Value)
}

// Match implements Expression interface.
func (e NotEqual) Match(payload map[string]interface{}) bool {
	return !reflect.DeepEqual(getField(payload, e.Field), e.Value)
}

// Match implements Expression interface.
func (e Exist) Match(payload map[string]interface{}) bool {
	_, found := getFieldExist(payload, e.Field)
	return found
}

// Match implements Expression interface.
func (e NotExist) Match(payload map[string]interface{}) bool {
	_, found := getFieldExist(payload, e.Field)
	return !found
}

// Match implements Expression interface.
func (e GreaterThan) Match(payload map[string]interface{}) bool {
	n, ok := isNumber(getField(payload, e.Field))
	return ok && (n > e.Value)
}

// Match implements Expression interface.
func (e GreaterOrEqual) Match(payload map[string]interface{}) bool {
	n, ok := isNumber(getField(payload, e.Field))
	return ok && (n >= e.Value)
}

// Match implements Expression interface.
func (e LowerThan) Match(payload map[string]interface{}) bool {
	n, ok := isNumber(getField(payload, e.Field))
	return ok && (n < e.Value)
}

// Match implements Expression interface.
func (e LowerOrEqual) Match(payload map[string]interface{}) bool {
	n, ok := isNumber(getField(payload, e.Field))
	return ok && (n <= e.Value)
}

// Match implements Expression interface.
func (e Regex) Match(payload map[string]interface{}) bool {
	return e.Value.MatchString(payload[e.Field].(string))
}
