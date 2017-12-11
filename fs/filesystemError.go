package fs

import "fmt"

type Error struct {
	Message string
}

var (
	ErrForbidden = Error{"You are not allowed to access this file"}
)

func (e Error) String() string {
	return fmt.Sprintf("[FS]%s", e.Message)
}

func (e Error) Error() string {
	return fmt.Sprintf("[FS]%s", e.Message)
}
