package schema

import "fmt"

// PreQuery stores a query as string so it can be parsed/validated later
type PreQuery struct {
	s string
	q Query
}

// Q is like ParseQuery, but returns an intermediate object which can
// be stored before behing parsed.
func Q(q string) *PreQuery {
	return &PreQuery{s: q}
}

func (q *PreQuery) compile(v Validator) error {
	if q.q == nil {
		var err error
		q.q, err = ParseQuery(q.s, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// compileDependencies recusively compiles all field.Dependency against the validator
// and report any error
func compileDependencies(s Schema, v Validator) error {
	for _, def := range s {
		if def.Dependency != nil {
			if err := def.Dependency.compile(v); err != nil {
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
		if field.Dependency != nil {
			if !field.Dependency.q.Match(doc) {
				addFieldError(errs, name, fmt.Sprintf("does not match dependency: %s", field.Dependency.s))
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
