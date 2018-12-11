package schema

// AllOf validates that all the sub field validators validates. Be aware that
// the order of the validators matter, as the result of one successful
// validation is passed as input to the next.
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

// ValidateQuery implements schema.FieldQueryValidator interface. Note the
// result of one successful validation is passed as input to the next. The
// result of the first error or last successful validation is returned.
func (v AllOf) ValidateQuery(value interface{}) (interface{}, error) {
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

// Validate ensures that all sub-validators validates. Note the result of one
// successful validation is passed as input to the next. The result of the first
// error or last successful validation is returned.
func (v AllOf) Validate(value interface{}) (interface{}, error) {
	for _, validator := range v {
		var err error
		if value, err = validator.Validate(value); err != nil {
			return nil, err
		}
	}
	return value, nil
}

// GetField implements the FieldGetter interface. Note that it will return the
// first matching field only.
func (v AllOf) GetField(name string) *Field {
	for _, obj := range v {
		if fg, ok := obj.(FieldGetter); ok {
			return fg.GetField(name)
		}
	}
	return nil
}
