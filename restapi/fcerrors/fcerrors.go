package fcerrors

import (
	"net/http"

	"github.com/freecloudio/server/models"
)

// Code is a specific error code, containing an error message
type Code struct {
	Msg        string
	StatusCode int
}

var (
	// Internal Server Error, this should only be used as a fallback
	Internal = Code{"Internal Server Error", http.StatusInternalServerError}
	// InvalidUserData is thrown when user data validation fails on signup
	InvalidUserData = Code{"Invalid user data", http.StatusBadRequest}
	// UserExists is thrown when a user already exists on signup
	UserExists = Code{"A user with the same email or username already exists", http.StatusBadRequest}
	// UserNotFound is pretty clear
	UserNotFound = Code{"User cannot be found", http.StatusNotFound}
	// HashingFailed is thrown when a password hash operation failed
	HashingFailed = Code{"Password hashing failed", http.StatusInternalServerError}
	// Database is thrown when a DB operation failed - Try to use more fine-grained errors
	Database = Code{"Database error", http.StatusInternalServerError}
	// Filesystem operation failed
	Filesystem = Code{"Filesystem error", http.StatusInternalServerError}
	// BadCredentials for login
	BadCredentials = Code{"Email or Password incorrect", http.StatusUnauthorized}
	// MissingCredentials from the request
	MissingCredentials = Code{"Email or Password are missing", http.StatusBadRequest}
	// DeleteSession failed
	DeleteSession = Code{"Could not delete session", http.StatusInternalServerError}
	// ExpiredSession provided
	ExpiredSession = Code{"Session is expired", http.StatusUnauthorized}
	// FileNotExists if the path could not be found in the db
	FileNotExists = Code{"File could not be found", http.StatusNotFound}
	// PathNotValid represents an path with bad characters
	PathNotValid = Code{"Path is not valid", http.StatusBadRequest}
)

// FCError is a struct implementing the Error interface, which should be used on all internal errors.
type FCError struct {
	// Message is a plain-text message that summarizes the error
	Message string
	// Code is used to associate a http return code with the given error
	Code Code
}

// New returns a new FCError with the given code and its default message
func New(code Code) error {
	return &FCError{Message: code.Msg, Code: code}
}

// NewMsg returns a new FCError with the given code and message
func NewMsg(code Code, message string) error {
	return &FCError{Message: message, Code: code}
}

// Wrap any error with the given code. Useful for wrapping database errors etc.
// If the given error is nil, nil will be returned
func Wrap(err error, code Code) error {
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
	return err.Code.StatusCode
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

// GetAPIError returns an Error model, as defined by the API
func GetAPIError(err error) *models.Error {
	return &models.Error{Message: err.Error()}
}
