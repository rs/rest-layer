package rest

var (
	// NotFoundError represents a 404 HTTP error.
	NotFoundError = &Error{404, "Not Found", nil}
	// PreconditionFailedError happends when a conditional request condition is not met.
	PreconditionFailedError = &Error{412, "Precondition Failed", nil}
	// ConflictError happens when another thread or node modified the data concurrently
	// with our own thread in such a way we can't securely apply the requested changes.
	ConflictError = &Error{409, "Conflict", nil}
	// InvalidMethodError happends when the used HTTP method is not supported for this
	// resource
	InvalidMethodError = &Error{405, "Invalid method", nil}
)

// Error defines a REST error
type Error struct {
	// Code defines the error code to be used for the error and for the HTTP status
	Code int
	// Message is the error message
	Message string
	// Issues holds per fields errors if any
	Issues map[string][]interface{}
}

func (e *Error) Error() string {
	return e.Message
}
