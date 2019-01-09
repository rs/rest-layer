package schema

// Compiler is similar to the Compiler interface, but intended for types that implements, or may hold, a
// reference. All nested types must implement this interface.
type Compiler interface {
	Compile(rc ReferenceChecker) error
}

// ReferenceChecker is used to retrieve a FieldValidator that can be used for validating referenced IDs.
type ReferenceChecker interface {
	// ReferenceChecker should return a FieldValidator that can be used for validate that a referenced ID exists and
	// is of the right format. If there is no resource matching path, nil should e returned.
	ReferenceChecker(path string) (FieldValidator, Validator)
}

// ReferenceCheckerFunc is an adapter that allows ordinary functions to be used as reference checkers.
type ReferenceCheckerFunc func(path string) FieldValidator

// ReferenceChecker calls f(path).
func (f ReferenceCheckerFunc) ReferenceChecker(path string) FieldValidator {
	return f(path)
}
