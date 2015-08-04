package schema

import (
	"errors"
	"fmt"
	"reflect"
)

// Field specifies the info for a single field of a spec
type Field struct {
	// Required throws an error when the field is not provided at creation.
	Required bool
	// ReadOnly throws an error when a field is changed by the client.
	// Default and OnInit/OnUpdate hooks can be used to set/change read-only
	// fields.
	ReadOnly bool
	// Default defines the value be stored on the field when when item is
	// created and this field is not provided by the client.
	Default interface{}
	// OnInit can be set to a function to generate the value of this field
	// when item is created. The function takes the current value if any
	// and returns the value to be stored.
	OnInit *func(value interface{}) interface{}
	// OnUpdate can be set to a function to generate the value of this field
	// when item is updated. The function takes the current value if any
	// and returns the value to be stored.
	OnUpdate *func(value interface{}) interface{}
	// Validator is used to validate the field's format.
	Validator FieldValidator
	// Filterable defines that the field can be used with the `filter` parameter.
	// When this property is set to `true`, you may want to ensure the backend
	// database has this field indexed.
	Filterable bool
	// Sortable defines that the field can be used with the `sort` paramter.
	// When this property is set to `true`, you may want to ensure the backend
	// database has this field indexed.
	Sortable bool
	// Schema can be set to a sub-schema to allow multi-level schema.
	Schema *Schema
}

// FieldValidator is an interface for all individual validators. It takes a value
// to validate as argument and returned the normalized value or an error if validation failed.
type FieldValidator interface {
	Validate(value interface{}) (interface{}, error)
}

// Compile implements Compiler interface and recusively compile sub schemas and validators
// when they implement Compiler interface
func (f Field) Compile() error {
	if f.Schema != nil {
		// Recusively compile sub schema if any
		if err := f.Schema.Compile(); err != nil {
			return fmt.Errorf(".%s", err.Error())
		}
	} else if f.Validator != nil {
		// Compile validator if it implements Compiler interface
		if c, ok := f.Validator.(Compiler); ok {
			if err := c.Compile(); err != nil {
				return fmt.Errorf(": %s", err.Error())
			}
		}
		if reflect.ValueOf(f.Validator).Kind() != reflect.Ptr {
			return errors.New(": not a schema.Validator pointer")
		}
	}
	return nil
}
