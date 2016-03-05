package schema

// Params defines a field params handler as well as validators for the supported params
// for the field.
type Params struct {
	// Defines the list of parameter names with their associated validators.
	Validators map[string]FieldValidator
	// Handler is the piece of logic modifying the parameter value based on passed parameters.
	Handler func(value interface{}, params map[string]interface{}) (interface{}, error)
}
