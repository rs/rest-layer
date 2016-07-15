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

// Compile implements Compiler interface
func (v *Object) Compile() error {
	if v.Schema == nil {
		return fmt.Errorf("No schema defined for object")
	}

	if err := compileDependencies(*v.Schema, v.Schema); err != nil {
		return err
	}
	return nil
}

// ErrorMap to return lots of errors
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

// Validate implements FieldValidator interface
func (v Object) Validate(value interface{}) (interface{}, error) {
	dict, ok := value.(map[string]interface{})
	if !ok {
		return nil, errors.New("not a dict")
	}
	dest, errs := v.Schema.Validate(nil, dict)
	if len(errs) > 0 {
		var errMap ErrorMap
		errMap = errs
		return nil, errMap
	}
	return dest, nil
}
