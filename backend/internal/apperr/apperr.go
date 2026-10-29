// Package apperr defines the domain error services and handlers return for
// expected failures and the codes the API error response carries.
package apperr

// Code is the machine-readable error code of the API error response.
type Code string

const (
	CodeValidation     Code = "validation_error"
	CodeAuthentication Code = "authentication_error"
	CodeForbidden      Code = "forbidden"
	CodeAuthorization  Code = "authorization_error"
	CodeNotFound       Code = "not_found"
	CodeConflict       Code = "conflict"
	CodeInternal       Code = "internal_error"
)

// Error is an expected failure with a code and a message that is safe to show
// to the user. Details is optional; Err wraps the cause for logging.
type Error struct {
	Code    Code
	Message string
	Details map[string]any
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}
