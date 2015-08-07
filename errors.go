package rest

import "golang.org/x/net/context"

var (
	// NotFoundError represents a 404 HTTP error.
	NotFoundError = &Error{404, "Not Found", nil}
	// PreconditionFailedError happends when a conditional request condition is not met.
	PreconditionFailedError = &Error{412, "Precondition Failed", nil}
	// ConflictError happens when another thread or node modified the data concurrently
	// with our own thread in such a way we can't securely apply the requested changes.
	ConflictError = &Error{409, "Conflict", nil}
	// InvalidMethodError happends when the used HTTP method is not supported for this
	// resource.
	InvalidMethodError = &Error{405, "Invalid Method", nil}
	// ClientClosedRequestError is returned when the client closed the connection before
	// the server was able to finish processing the request.
	ClientClosedRequestError = &Error{499, "Client Closed Request", nil}
	// NotImplementedError happends when a requested feature is not implemented.
	NotImplementedError = &Error{501, "Not Implemented", nil}
	// GatewayTimeoutError is returned when the specified timeout for the request has been
	// reached before the server was able to process it.
	GatewayTimeoutError = &Error{504, "Deadline Exceeded", nil}
	// UnknownError is thrown when the origine of the error can't be identified.
	UnknownError = &Error{520, "Unknown Error", nil}
)

// Error defines a REST error with optional per fields error details
type Error struct {
	// Code defines the error code to be used for the error and for the HTTP status
	Code int
	// Message is the error message
	Message string
	// Issues holds per fields errors if any
	Issues map[string][]interface{}
}

// ContextError takes a context.Context error returned by ctx.Err() and return the
// appropriate rest.Error.
//
// This method is to be used with `net/context` when the context's deadline is reached.
// Pass the output or `ctx.Err()` to this method to get the corresponding rest.Error.
func ContextError(err error) *Error {
	switch err {
	case context.Canceled:
		return ClientClosedRequestError
	case context.DeadlineExceeded:
		return GatewayTimeoutError
	case nil:
		return nil
	default:
		return &Error{520, err.Error(), nil}
	}
}

// Error returns the error as string
func (e *Error) Error() string {
	return e.Message
}
