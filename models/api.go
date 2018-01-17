package models

type APIError struct {
	Message string `json:"message,omitempty"`
	Code    int    `json:"-"`
}

func (err APIError) Error() string {
	return err.Message
}

// SuccessResponse is the minimal response to send if the request went OK
var SuccessResponse = struct {
	Success bool `json:"success"`
}{
	Success: true,
}
