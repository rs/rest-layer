package schema

// Params defines param name => definition pairs allowed for a field
type Params map[string]Param

// Param define an individual field parameter with its validator
type Param struct {
	// Description of the parameter
	Description string
	// Validator to use for this parameter
	Validator FieldValidator
}
