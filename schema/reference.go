package schema

import (
	"errors"
	"fmt"
)

// Reference validates the ID of a linked resource.
type Reference struct {
	Path            string
	validator       FieldValidator
	SchemaValidator Validator
}

// Compile validates v.Path against rc and stores the a FieldValidator for later use by v.Validate.
func (r *Reference) Compile(rc ReferenceChecker) error {
	if rc == nil {
		return fmt.Errorf("rc can not be nil")
	}

	if v, sv := rc.ReferenceChecker(r.Path); v != nil && sv != nil {
		r.validator = v
		r.SchemaValidator = sv
		return nil
	}

	return fmt.Errorf("can't find resource '%s'", r.Path)
}

// Validate validates and sanitizes IDs against the reference path.
func (r Reference) Validate(value interface{}) (interface{}, error) {
	if r.validator == nil {
		return nil, errors.New("not successfully compiled")
	}

	return r.validator.Validate(value)
}

// GetField implements the FieldGetter interface.
func (r Reference) GetField(name string) *Field {
	return r.SchemaValidator.GetField(name)
}
