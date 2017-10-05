package schema

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Object validates objects which are defined by Schemas.
type Object struct {
	Schema *Schema
}

// Compile implements the ReferenceCompiler interface.
func (v *Object) Compile(rc ReferenceChecker) error {
	if v.Schema == nil {
		return errors.New("no schema defined")
	}
	if err := compileDependencies(*v.Schema, v.Schema); err != nil {
		return err
	}
	return v.Schema.Compile(rc)
}

// Validate implements FieldValidator interface.
func (v Object) Validate(value interface{}) (interface{}, error) {
	obj, ok := value.(map[string]interface{})
	if !ok {
		return nil, errors.New("not an object")
	}
	dest, errs := v.Schema.Validate(nil, obj)
	if len(errs) > 0 {
		var errMap ErrorMap
		errMap = errs
		return nil, errMap
	}
	return dest, nil
}

// GetField implements the FieldGetter interface.
func (v Object) GetField(name string) *Field {
	return v.Schema.GetField(name)
}

// ErrorMap contains a map of errors by field name.
type ErrorMap map[string][]interface{}

func (e ErrorMap) Error() string {
	errs := make([]string, 0, len(e))
	for key := range e {
		errs = append(errs, key)
	}
	sort.Strings(errs)
	for i, key := range errs {
		errs[i] = fmt.Sprintf("%s is %s", key, e[key])
	}
	return strings.Join(errs, ", ")
}
