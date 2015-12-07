package schema

// Reference validates time based values
type Reference struct {
	Path string
}

// Validate validates and normalize reference based value
func (v Reference) Validate(value interface{}) (interface{}, error) {
	// All the work is performed in rest.checkReferences()
	return value, nil
}
