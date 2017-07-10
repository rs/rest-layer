package schema

// AllOf validates that all the sub field validators validates.
type AllOf []FieldValidator

// Compile implements the ReferenceCompiler interface.
func (v AllOf) Compile(rc ReferenceChecker) (err error) {
	for _, sv := range v {
		if c, ok := sv.(Compiler); ok {
			if err = c.Compile(rc); err != nil {
				return
			}
		}
	}
	return
}

// Validate ensures that all sub-validators validates.
func (v AllOf) Validate(value interface{}) (interface{}, error) {
	for _, validator := range v {
		var err error
		if value, err = validator.Validate(value); err != nil {
			return nil, err
		}
	}
	return value, nil
}
