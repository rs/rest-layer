package schema

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

// Fields defines a map of name -> field pairs
type Fields map[string]Field

// Field specifies the info for a single field of a spec
type Field struct {
	// Description stores a short description of the field useful for automatic
	// documentation generation.
	Description string
	// Required throws an error when the field is not provided at creation.
	Required bool
	// ReadOnly throws an error when a field is changed by the client.
	// Default and OnInit/OnUpdate hooks can be used to set/change read-only
	// fields.
	ReadOnly bool
	// Hidden allows writes but hides the field's content from the client. When
	// this field is enabled, PUTing the document without the field would not
	// remove the field but use the previous document's value if any.
	Hidden bool
	// Default defines the value be stored on the field when when item is
	// created and this field is not provided by the client.
	Default interface{}
	// OnInit can be set to a function to generate the value of this field
	// when item is created. The function takes the current value if any
	// and returns the value to be stored.
	OnInit func(ctx context.Context, value interface{}) interface{}
	// OnUpdate can be set to a function to generate the value of this field
	// when item is updated. The function takes the current value if any
	// and returns the value to be stored.
	OnUpdate func(ctx context.Context, value interface{}) interface{}
	// Params defines a param handler for the field. The handler may change the field's
	// value depending on the passed parameters.
	Params Params
	// Handler is the piece of logic modifying the field value based on passed parameters.
	// This handler is only called if at least on parameter is provided.
	Handler FieldHandler
	// Validator is used to validate the field's format. Please note you *must* pass in pointers to
	// FieldValidator instances otherwise `schema` will not be able to discover other interfaces,
	// such as `Compiler`, and *will* prevent schema from initializing specific FieldValidators
	// correctly causing unexpected runtime errors.
	// @see http://research.swtch.com/interfaces for more details.
	Validator FieldValidator
	// Dependency rejects the field if the schema predicate doesn't match the document.
	// Use query.MustParsePredicate(`{field: "value"}`) to populate this field.
	Dependency Predicate
	// Filterable defines that the field can be used with the `filter` parameter.
	// When this property is set to `true`, you may want to ensure the backend
	// database has this field indexed.
	Filterable bool
	// Sortable defines that the field can be used with the `sort` parameter.
	// When this property is set to `true`, you may want to ensure the backend
	// database has this field indexed.
	Sortable bool
	// Schema can be set to a sub-schema to allow multi-level schema.
	Schema *Schema
}

// Compile implements the ReferenceCompiler interface and recursively compile sub schemas
// and validators when they implement Compiler interface.
func (f Field) Compile(rc ReferenceChecker) error {
	// TODO check field name format (alpha num + _ and -).
	if f.Schema != nil {
		// Recursively compile sub schema if any.
		if err := f.Schema.Compile(rc); err != nil {
			return fmt.Errorf(".%v", err)
		}
	} else if f.Validator != nil {
		// Compile validator if it implements the ReferenceCompiler or Compiler interface.
		if c, ok := f.Validator.(Compiler); ok {
			if err := c.Compile(rc); err != nil {
				return fmt.Errorf(": %v", err)
			}
		}
		if reflect.ValueOf(f.Validator).Kind() != reflect.Ptr {
			return errors.New(": not a schema.Validator pointer")
		}
	}
	return nil
}

// FieldHandler is the piece of logic modifying the field value based on passed
// parameters
type FieldHandler func(ctx context.Context, value interface{}, params map[string]interface{}) (interface{}, error)

// FieldValidator is an interface for all individual validators. It takes a
// value to validate as argument and returned the normalized value or an error
// if validation failed.
type FieldValidator interface {
	Validate(value interface{}) (interface{}, error)
}

//FieldValidatorFunc is an adapter to allow the use of ordinary functions as
// field validators. If f is a function with the appropriate signature,
// FieldValidatorFunc(f) is a FieldValidator that calls f.
type FieldValidatorFunc func(value interface{}) (interface{}, error)

// Validate calls f(value).
func (f FieldValidatorFunc) Validate(value interface{}) (interface{}, error) {
	return f(value)
}

// FieldSerializer is used to convert the value between it's representation form
// and it internal storable form. A FieldValidator which implement this
// interface will have its Serialize method called before marshaling.
type FieldSerializer interface {
	// Serialize is called when the data is coming from it internal storable
	// form and needs to be prepared for representation (i.e.: just before JSON
	// marshaling).
	Serialize(value interface{}) (interface{}, error)
}

// FieldGetter defines an interface for fetching sub-fields from a Schema or
// FieldValidator implementation that allows (JSON) object values.
type FieldGetter interface {
	// GetField returns a Field for the given name if the name is allowed by
	// the schema. The field is expected to validate query values.
	//
	// You may reference a sub-field using dotted notation, e.g. field.subfield.
	GetField(name string) *Field
}

// LessFunc is a function that returns true only when value is less than other,
// and false in all other circumstances, including error conditions.
type LessFunc func(value, other interface{}) bool

// FieldComparator must be implemented by a FieldValidator that is to allow
// comparison queries ($gt, $gte, $lt and $lte). The returned LessFunc will be
// used by the query package's Predicate.Match functions, which is used e.g. by
// the internal mem storage backend.
type FieldComparator interface {
	// LessFunc returns a valid LessFunc or nil. nil is returned when comparison
	// is not allowed.
	LessFunc() LessFunc
}

// FieldQueryValidator defines an interface for lightweight validation on field
// types, without applying constrains on the actual values.
type FieldQueryValidator interface {
	ValidateQuery(value interface{}) (interface{}, error)
}
