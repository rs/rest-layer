package query

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/rs/rest-layer/schema"
)

const (
	opAnd            = "$and"
	opOr             = "$or"
	opExists         = "$exists"
	opIn             = "$in"
	opNotIn          = "$nin"
	opNotEqual       = "$ne"
	opLowerThan      = "$lt"
	opLowerOrEqual   = "$lte"
	opGreaterThan    = "$gt"
	opGreaterOrEqual = "$gte"
	opRegex          = "$regex"
)

// Predicate defines an expression against a schema to perform a match on schema's data.
type Predicate []Expression

// Match implements Expression interface.
func (e Predicate) Match(payload map[string]interface{}) bool {
	if e == nil || len(e) == 0 {
		// nil or empty predicates always match
		return true
	}
	// Run each sub queries like a root query, stop/pass on first match
	for _, subQuery := range e {
		if !subQuery.Match(payload) {
			return false
		}
	}
	return true
}

// String implements Expression interface.
func (e Predicate) String() string {
	if len(e) == 0 {
		return "{}"
	}
	s := make([]string, 0, len(e))
	for _, subQuery := range e {
		s = append(s, subQuery.String())
	}
	return "{" + strings.Join(s, ", ") + "}"
}

// Validate implements Expression interface.
func (e Predicate) Validate(validator schema.Validator) error {
	return validateExpressions(e, validator)
}

// Expression is a query or query component that can be matched against a payload.
type Expression interface {
	Match(payload map[string]interface{}) bool
	Validate(validator schema.Validator) error
	String() string
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

// String implements Expression interface.
func (e And) String() string {
	if len(e) == 0 {
		return opAnd + ": []"
	}
	s := make([]string, 0, len(e))
	for _, subQuery := range e {
		s = append(s, "{"+subQuery.String()+"}")
	}
	return opAnd + ": [" + strings.Join(s, ", ") + "]"
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

// String implements Expression interface.
func (e Or) String() string {
	if len(e) == 0 {
		return opOr + ": []"
	}
	s := make([]string, 0, len(e))
	for _, subQuery := range e {
		s = append(s, "{"+subQuery.String()+"}")
	}
	return opOr + ": [" + strings.Join(s, ", ") + "]"
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

// String implements Expression interface.
func (e In) String() string {
	s := make([]string, 0, len(e.Values))
	for _, v := range e.Values {
		s = append(s, valueString(v))
	}
	return quoteField(e.Field) + ": {" + opIn + ": [" + strings.Join(s, ", ") + "]}"
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

// String implements Expression interface.
func (e NotIn) String() string {
	s := make([]string, 0, len(e.Values))
	for _, v := range e.Values {
		s = append(s, valueString(v))
	}
	return quoteField(e.Field) + ": {" + opNotIn + ": [" + strings.Join(s, ", ") + "]}"
}

// Validate implements Expression interface.
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

// String implements Expression interface.
func (e Equal) String() string {
	return quoteField(e.Field) + ": " + valueString(e.Value)
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

// String implements Expression interface.
func (e NotEqual) String() string {
	return quoteField(e.Field) + ": {" + opNotEqual + ": " + valueString(e.Value) + "}"
}

// Exist matches all values which are present, even if nil
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

// String implements Expression interface.
func (e Exist) String() string {
	return quoteField(e.Field) + ": {" + opExists + ": true}"
}

// NotExist matches all values which are absent
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

// String implements Expression interface.
func (e NotExist) String() string {
	return quoteField(e.Field) + ": {" + opExists + ": false}"
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
	return validateNumericValue(e.Field, e.Value, opGreaterThan, validator)
}

// String implements Expression interface.
func (e GreaterThan) String() string {
	return quoteField(e.Field) + ": {" + opGreaterThan + ": " + valueString(e.Value) + "}"
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
	return validateNumericValue(e.Field, e.Value, opGreaterOrEqual, validator)
}

// String implements Expression interface.
func (e GreaterOrEqual) String() string {
	return quoteField(e.Field) + ": {" + opGreaterOrEqual + ": " + valueString(e.Value) + "}"
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
	return validateNumericValue(e.Field, e.Value, opLowerThan, validator)
}

// String implements Expression interface.
func (e LowerThan) String() string {
	return quoteField(e.Field) + ": {" + opLowerThan + ": " + valueString(e.Value) + "}"
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
	return validateNumericValue(e.Field, e.Value, opLowerOrEqual, validator)
}

// String implements Expression interface.
func (e LowerOrEqual) String() string {
	return quoteField(e.Field) + ": {" + opLowerOrEqual + ": " + valueString(e.Value) + "}"
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

// String implements Expression interface.
func (e Regex) String() string {
	return quoteField(e.Field) + ": {" + opRegex + ": " + valueString(e.Value) + "}"
}
