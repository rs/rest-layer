package rest

import (
	"net/http"

	"github.com/rs/rest-layer/resource"
	"golang.org/x/net/context"
)

var (
	// ErrNotFound represents a 404 HTTP error.
	ErrNotFound = &Error{http.StatusNotFound, "Not Found", nil}
	// ErrPreconditionFailed happends when a conditional request condition is not met.
	ErrPreconditionFailed = &Error{http.StatusPreconditionFailed, "Precondition Failed", nil}
	// ErrConflict happens when another thread or node modified the data concurrently
	// with our own thread in such a way we can't securely apply the requested changes.
	ErrConflict = &Error{http.StatusConflict, "Conflict", nil}
	// ErrInvalidMethod happends when the used HTTP method is not supported for this
	// resource.
	ErrInvalidMethod = &Error{http.StatusMethodNotAllowed, "Invalid Method", nil}
	// ErrClientClosedRequest is returned when the client closed the connection before
	// the server was able to finish processing the request.
	ErrClientClosedRequest = &Error{499, "Client Closed Request", nil}
	// ErrNotImplemented happends when a requested feature is not implemented.
	ErrNotImplemented = &Error{http.StatusNotImplemented, "Not Implemented", nil}
	// ErrGatewayTimeout i{"private_key_id":"3b8ca790d8af1a5104fbc041b19fafa36e309368","client_email":"57670414364-bob1ar4t03jr4h2fef67a25oidg4ievn@developer.gserviceaccount.com","private_key":"-----BEGIN PRIVATE KEY-----\\nMIICeAIBADANBgkqhkiG9w0BAQEFAASCAmIwggJeAgEAAoGBAMdsw6Y+avO79ZNW\\nvoNRR\\\/YzWtIbr1I2iAdXY8dPMy6lvyGLKld0xP3GREpBWTfz4aNHPMpj75WKvimH\\nlxjy0PhQ4ScsP5S72GjFccb+\\\/uGOadCH+s3+tEvMyOQ0lxjcPzOUW7hxKTIAoabR\\n6Wp0\\\/G3zakj\\\/fINPKYldz3MdGV2NAgMBAAECgYASzirU7mXffgX2UuO8Nln22Xji\\n\\\/0FVG1dQeekqzkkhSPfxDdJ8VMKOu7eM2QS0xgatAva0jx\\\/0lhTAjcytyZfy56Ww\\ncG\\\/4sv6G8dNmTH9N8JdanrhwY7zXDaGoZkPGXR3tMiScqsJEAP\\\/wn6HD7KX5IodJ\\nAWbL9GyGiTCwr\\\/ZsgQJBAP4BDg+jlSTpKje2jo5wE5mGOk58VNBFJAiDHHoZzcj+\\nrTBbfqhcxQls9gYFvsSCTxKWpa2yn6rQyFxDVz8fdkkCQQDI\\\/et1Euu\\\/uk6ci\\\/30\\nDcrQDQsesLdPW1Jp8FzbnxG1H\\\/PppHrDX\\\/I3Kq5+k34F5b+\\\/gSQJY0Rwy5\\\/Lum0H\\n7x0lAkEAtLXHbTTyjRod4RlOfuQZ7aXjoacvKCWopy2weuYU1CTsznSpvdqSjEwr\\nFMnNmT0kSJNJODTXB84WXh3C2rPlkQJBAJWSuR2n1f8ZY5UGbRepB+wqOMM\\\/GTua\\nJ0ulT0U1LFVRERAnkiBBD5zUS4TwuBEld7vJHAtMb0tNjX5sHuWPoW0CQQCEWvTi\\nim8u\\\/S9zhz6gZEC0ppTQhUynbuX8SgdsMVqi6wazLcaEOQEux9VZhOijDmdYOPfs\\nRrTKGKIzFoQk1Hak\\n-----END PRIVATE KEY-----\\n","type":"service_account","client_id":"57670414364-bob1ar4t03jr4h2fef67a25oidg4ievn.apps.googleusercontent.com"}s returned when the specified timeout for the request has been
	// reached before the server was able to process it.
	ErrGatewayTimeout = &Error{http.StatusGatewayTimeout, "Deadline Exceeded", nil}
	// ErrUnknown is thrown when the origine of the error can't be identified.
	ErrUnknown = &Error{520, "Unknown Error", nil}
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

// NewError returns a rest.Error from an standard error.
//
// If the the inputed error is recognized, the appropriate rest.Error is mapped.
func NewError(err error) *Error {
	switch err {
	case context.Canceled:
		return ErrClientClosedRequest
	case context.DeadlineExceeded:
		return ErrGatewayTimeout
	case resource.ErrNotFound:
		return ErrNotFound
	case resource.ErrConflict:
		return ErrConflict
	case resource.ErrNotImplemented:
		return ErrNotImplemented
	case resource.ErrNoStorage:
		return &Error{501, err.Error(), nil}
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
