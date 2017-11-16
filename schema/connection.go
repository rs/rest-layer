package schema

// Connection is a dummy validator to define a weak connection to another
// schema. The query.Projection will treat this validator as an external
// resource, and generate a sub-request to fetch the sub-payload.
type Connection struct {
	Path  string
	Field string
}

// Validate implements the FieldValidator interface.
func (v *Connection) Validate(value interface{}) (interface{}, error) {
	// No validation perform at this time.
	return value, nil
}
