package fcerrors

import "net/http"

// Code is a specific error code
type Code string

const (
	// Internal Server Error, this should only be used as a fallback
	Internal = Code("Internal")
	// InvalidUserData is thrown when user data validation fails on signup
	InvalidUserData = Code("InvalidUserData")
	// UserExists is thrown when a user already exists on signup
	UserExists = Code("UserExists")
	// UserNotFound is pretty clear
	UserNotFound = Code("UserNotFound")
	// HashingFailed is thrown when a password hash operation failed
	HashingFailed = Code("HashingFailed")
	// Database is thrown when a DB operation failed
	Database = Code("Database")
	// Filesystem operation failed
	Filesystem = Code("Filesystem")
	// BadCredentials for login
	BadCredentials = Code("BadCredentials")
	// MissingCredentials from the request
	MissingCredentials = Code("MissingData")
	// DeleteSession failed
	DeleteSession = Code("DeleteSession")
)

// FCError is a struct implementing the Error interface, which should be used on all internal errors.
type FCError struct {
	// Message is a plain-text message that summarizes the error
	Message string
	// Code is used to associate a http return code with the given error
	Code Code
}

// New returns a new FCError with the given message and code
func New(message string, code Code) *FCError {
	return &FCError{message, code}
}

// Wrap any error with the given code. Useful for wrapping database errors etc.
// If the given error is nil, nil will be returned
func Wrap(err error, code Code) *FCError {
	if err == nil {
		return nil
	}
	return &FCError{err.Error(), code}
}

func (err *FCError) Error() string {
	return err.Message
}

// GetStatusCode returns the HTTP status code associated with the given error, or 500 if it is unknown
func (err *FCError) GetStatusCode() int {
	switch err.Code {
	case Internal:
		return http.StatusInternalServerError
	case InvalidUserData:
		return http.StatusBadRequest
	case UserExists:
		return http.StatusBadRequest
	case UserNotFound:
		return http.StatusNotFound
	case HashingFailed:
		return http.StatusInternalServerError
	case Database:
		return http.StatusInternalServerError
	case Filesystem:
		return http.StatusInternalServerError
	case BadCredentials:
		return http.StatusUnauthorized
	case MissingCredentials:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// GetStatusCode returns the HTTP status code associated with the given error
// In case the given error is an FCError, the status code will be it associated one,
// otherwise, an internal server error will be returned.
func GetStatusCode(err error) int {
	if e, ok := err.(*FCError); ok {
		return e.GetStatusCode()
	}
	return http.StatusInternalServerError
}
