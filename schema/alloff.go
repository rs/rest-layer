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

// ValidateQuery implements schema.FieldQueryValidator interface
func (v AllOf) ValidateQuery(value interface{}) (interface{}, error) {
	// This works like this:
	// 1.	only the first validator gets passed the original value.
	// 2. after that, each validator gets passed the value from the previous validator.
	// 3. finally, the value returned from the last validator is returned for further use.
	for _, validator := range v {
		var err error
		if validatorQuery, ok := validator.(FieldQueryValidator); ok {
			value, err = validatorQuery.ValidateQuery(value)
		} else {
			value, err = validator.Validate(value)
		}
		if err != nil {
			return nil, err
		}
	}
	return value, nil
}

// Validate ensures that all sub-validators validates.
func (v AllOf) Validate(value interface{}) (interface{}, error) {
	// This works like this:
	// 1.	only the first validator gets passed the original value.
	// 2. after that, each validator gets passed the value from the previous validator.
	// 3. finally, the value returned from the last validator is returned for further use.
	for _, validator := range v {
		var err error
		if value, err = validator.Validate(value); err != nil {
			return nil, err
		}
	}
	return value, nil
}
