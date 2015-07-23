package schema

// AllOf validates that all the sub field validators validates
type AllOf []FieldValidator

// Validate ensures that all sub-validators validates
func (v AllOf) Validate(value interface{}) (interface{}, error) {
	for _, validator := range v {
		var err error
		if value, err = validator.Validate(value); err != nil {
			return nil, err
		}
	}
	return value, nil
}
