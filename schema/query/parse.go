package query

import (
	"errors"
	"fmt"
	"regexp"
)

// parse parses a Query from a map
func parse(q map[string]interface{}, parentKey string) (Query, error) {
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
			vals, ok := exp.([]interface{})
			if !ok {
				vals = []interface{}{exp}
			}
			for _, v := range vals {
				values = append(values, v)
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
				query, err := parse(sq, "")
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
			if parentKey != "" {
				return nil, fmt.Errorf("%s: invalid expression", parentKey)
			}
			if subQuery, ok := exp.(map[string]interface{}); ok {
				sq, err := parse(subQuery, key)
				if err != nil {
					return nil, err
				}
				queries = append(queries, sq...)
			} else {
				// Exact match
				queries = append(queries, Equal{Field: key, Value: exp})
			}
		}
	}
	return queries, nil
}
