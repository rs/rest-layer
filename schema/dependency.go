package schema

import (
	"fmt"
)

// Query is an interface matching the query.Query type.
type Query interface {
	Match(payload map[string]interface{}) bool
	Validate(v Validator) error
}

// Q is deprecated, use query.MustParse instead.
func Q() Query {
	panic("schema.Q is deprecated, please use query.MustParse instead")
}

// compileDependencies recursively compiles all field.Dependency against the
// validator and report any error.
func compileDependencies(s Schema, v Validator) error {
	for _, def := range s.Fields {
		if def.Dependency != nil {
			if err := def.Dependency.Validate(v); err != nil {
				return err
			}
		}
		if def.Schema != nil {
			if err := compileDependencies(*def.Schema, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s Schema) validateDependencies(changes map[string]interface{}, doc map[string]interface{}, prefix string) (errs map[string][]interface{}) {
	errs = map[string][]interface{}{}
	for name, value := range changes {
		path := prefix + name
		field := s.GetField(path)
		if field != nil && field.Dependency != nil {
			if !field.Dependency.Match(doc) {
				addFieldError(errs, name, fmt.Sprintf("does not match dependency: %+v", field.Dependency))
			}
		}
		if subChanges, ok := value.(map[string]interface{}); ok {
			if subErrs := s.validateDependencies(subChanges, doc, path+"."); len(subErrs) > 0 {
				addFieldError(errs, name, subErrs)
			}
		}
	}
	return errs
}
