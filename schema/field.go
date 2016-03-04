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
	OnInit *func(value interface{}, params []interface{}) interface{}
	// OnUpdate can be set to a function to generate the value of this field
	// when item is updated. The function takes the current value if any
	// and returns the value to be stored.
	OnUpdate *func(value interface{}, params []interface{}) interface{} 
	// HookParams define the params to be passed to OnInit and OnUpdate hooks.
	HookParams []HookParam
	// Params defines a param handler for the field. The handler may change the field's
	// value depending on the passed parameters.
	Params *Params
	// Validator is used to validate the field's format.
	Validator FieldValidator
	// Dependency rejects the field if the schema query doesn't match the document.
	// Use schema.Q(`{"field": "value"}`) to populate this field.
	Dependency *PreQuery
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

// Types of Hook Parameters
// - ConstValue: To be defined as constant value.
// - FieldValue: To be resolved with the value of a current field.
type HookParamType int
const (
	ConstValue HookParamType = iota
	FieldValue
)

// HookParam has all the information to generate the param vakue
type HookParam struct {
	Type HookParamType
	Param interface{}
}

// Prepare implements the resolution of a parameter depending on the Type.
func (h HookParam) Prepare(payload map[string]interface{}) interface{} {
	switch h.Type {
		case ConstValue:
			return h.Param
		case FieldValue:
			field, ok := h.Param.(string)
			if ok {
				return payload[field]
			}
	}
	return nil
}


// FieldValidator is an interface for all individual validators. It takes a value
// to validate as argument and returned the normalized value or an error if validation failed.
type FieldValidator interface {
	Validate(value interface{}) (interface{}, error)
}

// FieldSerializer is used to convert the value between it's representation form and it
// internal storable form. A FieldValidator which implement this interface will have its
// Serialize method called before marshaling.
type FieldSerializer interface {
	// Serialize is called when the data is comming from it internal storable form and
	// needs to be prepared for representation (i.e.: just before JSON marshaling)
	Serialize(value interface{}) (interface{}, error)
}

// Compile implements Compiler interface and recusively compile sub schemas and validators
// when they implement Compiler interface
func (f Field) Compile() error {
	// TODO check field name format (alpha num + _ and -)
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
