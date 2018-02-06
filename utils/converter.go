package utils

import (
	"path/filepath"
	"strings"
)

// ConvertToSlash takes a string (most likely a path) and converts all back-slashes to normal slashes
func ConvertToSlash(path string) string {
	cleanPath := filepath.Clean(path)
	return strings.Replace(cleanPath, "\\", "/", -1)
}
