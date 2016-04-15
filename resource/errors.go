package resource

import "errors"

var (
	// ErrNotFound is returned when the requested resource can't be found
	ErrNotFound = errors.New("Not Found")
	// ErrUnauthorized is returned when the requested resource can be accessed by the
	// requestor for security reason
	ErrUnauthorized = errors.New("Unauthorized")
	// ErrConflict happens when another thread or node modified the data concurrently
	// with our own thread in such a way we can't securely apply the requested changes.
	ErrConflict = errors.New("Conflict")
	// ErrNotImplemented happens when a used filter is not implemented by the storage
	// handler.
	ErrNotImplemented = errors.New("Not Implemented")
	// ErrNoStorage is returned when not storage handler has been set on the resource.
	ErrNoStorage = errors.New("No Storage Defined")
)
