package schema

// AnyOf validates if any of the sub field validators validates. If any of the
// sub field validators implements the FieldSerializer interface, the *first*
// implementation which does not error will be used.
type AnyOf []FieldValidator

// Compile implements the Compiler interface.
func (v AnyOf) Compile(rc ReferenceChecker) error {
	for _, sv := range v {
		if c, ok := sv.(Compiler); ok {
			if err := c.Compile(rc); err != nil {
				return err
			}
		}

	}
	return nil
}

// ValidateQuery implements schema.FieldQueryValidator interface.
func (v AnyOf) ValidateQuery(value interface{}) (interface{}, error) {
	var errs ErrorSlice

	for _, validator := range v {
		var err error
		var val interface{}
		if validatorQuery, ok := validator.(FieldQueryValidator); ok {
			val, err = validatorQuery.ValidateQuery(value)
		} else {
			val, err = validator.Validate(value)
		}
		if err == nil {
			return val, nil
		}
		errs = errs.Append(err)
	}

	if len(errs) > 0 {
		return nil, errs
	}
	return nil, nil
}

// Validate ensures that at least one sub-validator validates.
func (v AnyOf) Validate(value interface{}) (interface{}, error) {
	var errs ErrorSlice

	for _, validator := range v {
		value, err := validator.Validate(value)
		if err == nil {
			return value, nil
		}
		errs = errs.Append(err)
	}

	if len(errs) > 0 {
		return nil, errs
	}
	return nil, nil
}

// Serialize attempts to serialize the value using the first available
// FieldSerializer which does not return an error. If no appropriate serializer
// is found, the input value is returned.
func (v AnyOf) Serialize(value interface{}) (interface{}, error) {
	for _, serializer := range v {
		s, ok := serializer.(FieldSerializer)
		if !ok {
			continue
		}

		v, err := s.Serialize(value)
		if err != nil {
			continue
		}
		return v, nil
	}

	return value, nil
}

// LessFunc implements the FieldComparator interface, and returns the first
// non-nil LessFunc or nil.
func (v AnyOf) LessFunc() LessFunc {
	for _, comparable := range v {
		if fc, ok := comparable.(FieldComparator); ok {
			if less := fc.LessFunc(); less != nil {
				return less
			}
		}
	}
	return nil
}

// GetField implements the FieldGetter interface. Note that it will return the
// first matching field only.
func (v AnyOf) GetField(name string) *Field {
	for _, obj := range v {
		if fg, ok := obj.(FieldGetter); ok {
			return fg.GetField(name)
		}
	}
	return nil
}
