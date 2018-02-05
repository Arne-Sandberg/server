package utils

import (
	"strings"
)

// ConvertToSlash takes a string (most likely a path) and converts all back-slashes to normal slashes
func ConvertToSlash(path string) string {
	return strings.Replace(path, "\\", "/", -1)
}
