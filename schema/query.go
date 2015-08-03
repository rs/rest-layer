package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// Query defines an expression against a schema to perform a match schema's data
type Query map[string]interface{}

// NewQuery returns a new query with the provided key/value
func NewQuery(q map[string]interface{}) Query {
	nq := Query{}
	for key, exp := range q {
		nq[key] = exp
	}
	return nq
}

// ParseQuery parses and validate a query as string
func ParseQuery(query string, validator Validator) (Query, error) {
	var j interface{}
	json.Unmarshal([]byte(query), &j)
	q, ok := j.(map[string]interface{})
	if !ok {
		return nil, errors.New("must be a JSON object")
	}
	if err := validateQuery(q, validator, ""); err != nil {
		return nil, err
	}
	return q, nil
}

// validateQuery recursively validates the format of a query
func validateQuery(q map[string]interface{}, validator Validator, parentKey string) error {
	for key, exp := range q {
		switch key {
		case "$ne":
			op := key
			if parentKey == "" {
				return fmt.Errorf("%s can't be at first level", op)
			}
			if field := validator.GetField(parentKey); field != nil {
				if field.Validator != nil {
					if _, err := field.Validator.Validate(exp); err != nil {
						return fmt.Errorf("invalid query expression for field `%s': %s", parentKey, err)
					}
				}
			} else {
				return fmt.Errorf("unknown query field: %s", parentKey)
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
					case Integer, Float:
						if _, err := field.Validator.Validate(exp); err != nil {
							return fmt.Errorf("invalid query expression for field `%s': %s", parentKey, err)
						}
					default:
						return fmt.Errorf("%s: cannot apply %s operation on a non numerical field", parentKey, op)
					}
				}
			} else {
				return fmt.Errorf("unknown query field: %s", parentKey)
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
							return fmt.Errorf("invalid query expression (%s) for field `%s': %s", value, parentKey, err)
						}
					}
				}
			} else {
				return fmt.Errorf("unknown query field: %s", parentKey)
			}
		case "$or":
			var subQueries []interface{}
			var ok bool
			if subQueries, ok = exp.([]interface{}); !ok {
				return errors.New("value for $or must be an array of dicts")
			}
			if len(subQueries) < 2 {
				return errors.New("$or must contain at least to elements")
			}
			// Cast map to Query object
			castedExp := make([]Query, len(subQueries))
			for i, subQuery := range subQueries {
				sq, ok := subQuery.(map[string]interface{})
				if !ok {
					return errors.New("value for $or must be an array of dicts")
				} else if err := validateQuery(sq, validator, ""); err != nil {
					return err
				}
				castedExp[i] = sq
			}
			q[key] = castedExp
		default:
			// Exact match
			if parentKey != "" {
				return fmt.Errorf("%s: invalid expression", parentKey)
			}
			if subQuery, ok := exp.(map[string]interface{}); ok {
				if err := validateQuery(subQuery, validator, key); err != nil {
					return err
				}
				// Cast map to Query object
				q[key] = Query(subQuery)
			} else {
				if field := validator.GetField(key); field != nil {
					if field.Validator != nil {
						if _, err := field.Validator.Validate(exp); err != nil {
							return fmt.Errorf("invalid query expression for field `%s': %s", key, err)
						}
					}
				} else {
					return fmt.Errorf("unknown query field: %s", key)
				}
			}
		}
	}
	return nil
}

// Match executes the query on the given payload and tells if it match
func (q Query) Match(payload map[string]interface{}) bool {
	return matchQuery(q, payload, "")
}

func matchQuery(q Query, payload map[string]interface{}, parentKey string) bool {
	for key, exp := range q {
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
			if subQueries, ok := exp.([]Query); ok {
				// Run each sub queries like a root query, stop/pass on first match
				for _, subQuery := range subQueries {
					if matchQuery(subQuery, payload, "") {
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
			if subQuery, ok := exp.(Query); ok {
				if !matchQuery(subQuery, payload, key) {
					return false
				}
			} else if !reflect.DeepEqual(getField(payload, key), exp) {
				return false
			}
		}
	}
	return true
}
