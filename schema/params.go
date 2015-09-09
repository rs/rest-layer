package schema

// Params defines a field params handler as well as validators for the supported params
// for the field.
type Params struct {
	Handler    func(value interface{}, params map[string]interface{}) (interface{}, error)
	Validators map[string]FieldValidator
}
