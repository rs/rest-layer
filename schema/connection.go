package schema

// Connection is a dummy validator to define a weak connection to another
// schema. The query.Projection will treat this validator as an external
// resource, and generate a sub-request to fetch the sub-payload.
type Connection struct {
	Path string
}

func (v *Connection) Compile(rc ReferenceChecker) (err error) {
	// Nothing to compile, implemented to force Connection on pointer.
	return nil
}

func (v Connection) Validate(value interface{}) (interface{}, error) {
	// No validation perform at this time.
	return value, nil
}
