package query

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/rs/rest-layer/schema"
)

// Query defines an expression against a schema to perform a match on schema's data.
type Query []Expression

// Parse parses a query.
func Parse(query string) (Query, error) {
	var j interface{}
	if err := json.Unmarshal([]byte(query), &j); err != nil {
		return nil, errors.New("must be valid JSON")
	}
	q, ok := j.(map[string]interface{})
	if !ok {
		return nil, errors.New("must be a JSON object")
	}
	return parse(q, "")
}

func MustParse(query string) Query {
	q, err := Parse(query)
	if err != nil {
		panic(fmt.Sprintf("query: Parse(%q): %v", query, err))
	}
	return q
}

// Match implements Expression interface.
func (e Query) Match(payload map[string]interface{}) bool {
	// Run each sub queries like a root query, stop/pass on first match
	for _, subQuery := range e {
		if !subQuery.Match(payload) {
			return false
		}
	}
	return true
}

// Validate implements Expression interface.
func (e Query) Validate(validator schema.Validator) error {
	return validateExpressions(e, validator)
}

// Expression is a query or query component that can be matched against a payload.
type Expression interface {
	Match(payload map[string]interface{}) bool
	Validate(validator schema.Validator) error
}

// Value represents any kind of value to use in query.
type Value interface{}

// And joins query clauses with a logical AND, returns all documents that match
// the conditions of both clauses.
type And []Expression

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

// Validate implements Expression interface.
func (e And) Validate(validator schema.Validator) error {
	return validateExpressions(e, validator)
}

// Or joins query clauses with a logical OR, returns all documents that
// match the conditions of either clause.
type Or []Expression

// Match implements Expression interface.
func (e Or) Match(payload map[string]interface{}) bool {
	// Run each sub queries like a root query, stop/pass on first match
	for _, subQuery := range e {
		if subQuery.Match(payload) {
			return true
		}
	}
	return false
}

// Validate implements Expression interface.
func (e Or) Validate(validator schema.Validator) error {
	return validateExpressions(e, validator)
}

// In matches any of the values specified in an array.
type In struct {
	Field  string
	Values []Value
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

// Validate implements Expression interface.
func (e In) Validate(validator schema.Validator) error {
	return validateValues(e.Field, e.Values, validator)
}

// NotIn matches none of the values specified in an array.
type NotIn struct {
	Field  string
	Values []Value
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

// Validate implements Expression interface..
func (e NotIn) Validate(validator schema.Validator) error {
	return validateValues(e.Field, e.Values, validator)
}

// Equal matches all values that are equal to a specified value.
type Equal struct {
	Field string
	Value Value
}

// Match implements Expression interface.
func (e Equal) Match(payload map[string]interface{}) bool {
	return reflect.DeepEqual(getField(payload, e.Field), e.Value)
}

// Validate implements Expression interface.
func (e Equal) Validate(validator schema.Validator) error {
	return validateValue(e.Field, e.Value, validator)
}

// NotEqual matches all values that are not equal to a specified value.
type NotEqual struct {
	Field string
	Value Value
}

// Match implements Expression interface.
func (e NotEqual) Match(payload map[string]interface{}) bool {
	return !reflect.DeepEqual(getField(payload, e.Field), e.Value)
}

// Validate implements Expression interface.
func (e NotEqual) Validate(validator schema.Validator) error {
	return validateValue(e.Field, e.Value, validator)
}

// Exist matches all values which are present, even if nil.
type Exist struct {
	Field string
}

// Match implements Expression interface.
func (e Exist) Match(payload map[string]interface{}) bool {
	_, found := getFieldExist(payload, e.Field)
	return found
}

// Validate implements Expression interface.
func (e Exist) Validate(validator schema.Validator) error {
	return validateField(e.Field, validator)
}

// NotExist matches all values which are absent.
type NotExist struct {
	Field string
}

// Match implements Expression interface.
func (e NotExist) Match(payload map[string]interface{}) bool {
	_, found := getFieldExist(payload, e.Field)
	return !found
}

// Validate implements Expression interface.
func (e NotExist) Validate(validator schema.Validator) error {
	return validateField(e.Field, validator)
}

// GreaterThan matches values that are greater than a specified value.
type GreaterThan struct {
	Field string
	Value float64
}

// Match implements Expression interface.
func (e GreaterThan) Match(payload map[string]interface{}) bool {
	n, ok := isNumber(getField(payload, e.Field))
	return ok && (n > e.Value)
}

// Validate implements Expression interface.
func (e GreaterThan) Validate(validator schema.Validator) error {
	return validateNumericValue(e.Field, e.Value, "$gt", validator)
}

// GreaterOrEqual matches values that are greater than or equal to a specified value.
type GreaterOrEqual struct {
	Field string
	Value float64
}

// Match implements Expression interface
func (e GreaterOrEqual) Match(payload map[string]interface{}) bool {
	n, ok := isNumber(getField(payload, e.Field))
	return ok && (n >= e.Value)
}

// Validate implements Expression interface.
func (e GreaterOrEqual) Validate(validator schema.Validator) error {
	return validateNumericValue(e.Field, e.Value, "$ge", validator)
}

// LowerThan matches values that are less than a specified value.
type LowerThan struct {
	Field string
	Value float64
}

// Match implements Expression interface.
func (e LowerThan) Match(payload map[string]interface{}) bool {
	n, ok := isNumber(getField(payload, e.Field))
	return ok && (n < e.Value)
}

// Validate implements Expression interface.
func (e LowerThan) Validate(validator schema.Validator) error {
	return validateNumericValue(e.Field, e.Value, "$lt", validator)
}

// LowerOrEqual matches values that are less than or equal to a specified value.
type LowerOrEqual struct {
	Field string
	Value float64
}

// Match implements Expression interface.
func (e LowerOrEqual) Match(payload map[string]interface{}) bool {
	n, ok := isNumber(getField(payload, e.Field))
	return ok && (n <= e.Value)
}

// Validate implements Expression interface.
func (e LowerOrEqual) Validate(validator schema.Validator) error {
	return validateNumericValue(e.Field, e.Value, "$le", validator)
}

// Regex matches values that match to a specified regular expression.
type Regex struct {
	Field string
	Value *regexp.Regexp
}

// Match implements Expression interface.
func (e Regex) Match(payload map[string]interface{}) bool {
	return e.Value.MatchString(payload[e.Field].(string))
}

// Validate implements Expression interface.
func (e Regex) Validate(validator schema.Validator) error {
	return validateValue(e.Field, e.Value, validator)
}
